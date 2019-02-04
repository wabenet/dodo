package image

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"

	"github.com/moby/buildkit/session"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
)

func prepareSession(baseDir string) (*session.Session, error) {
	sessionID, err := readOrCreateSessionID()
	if err != nil {
		return nil, err
	}

	s := sha256.Sum256([]byte(fmt.Sprintf("%s:%s", sessionID, baseDir)))
	sharedKey := hex.EncodeToString(s[:])

	session, err := session.NewSession(context.Background(), filepath.Base(baseDir), sharedKey)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create session")
	}

	return session, nil
}

func readOrCreateSessionID() (string, error) {
	user, err := user.Current()
	if err != nil {
		return "", err
	}

	sessionDir := filepath.Join(user.HomeDir, ".dodo")
	sessionFile := filepath.Join(sessionDir, "sessionID")

	err = os.MkdirAll(sessionDir, 0700)
	if err != nil {
		return "", err
	}

	if _, err = os.Lstat(sessionFile); err == nil {
		sessionID, err := ioutil.ReadFile(sessionFile)
		if err != nil {
			return "", err
		}
		return string(sessionID), nil
	}

	sessionID := make([]byte, 32)
	_, err = rand.Read(sessionID)
	if err != nil {
		return "", err
	}
	sessionID = []byte(hex.EncodeToString(sessionID))
	err = ioutil.WriteFile(sessionFile, sessionID, 0600)
	if err != nil {
		return "", err
	}

	return string(sessionID), nil
}
