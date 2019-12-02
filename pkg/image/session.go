package image

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	buildkit "github.com/moby/buildkit/session"
	"github.com/oclaussen/dodo/pkg/config"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
)

type session interface {
	ID() string
	Allow(buildkit.Attachable)
	Run(context.Context, buildkit.Dialer) error
	Close() error
}

func prepareSession(baseDir string) (session, error) {
	sessionID, err := readOrCreateSessionID()
	if err != nil {
		return nil, err
	}

	s := sha256.Sum256([]byte(fmt.Sprintf("%s:%s", sessionID, baseDir)))
	sharedKey := hex.EncodeToString(s[:])

	session, err := buildkit.NewSession(context.Background(), filepath.Base(baseDir), sharedKey)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create session")
	}

	return session, nil
}

func readOrCreateSessionID() (string, error) {
	sessionFile := filepath.Join(config.GetAppDir(), "sessionID")
	if _, err := os.Lstat(sessionFile); err == nil {
		sessionID, err := ioutil.ReadFile(sessionFile)
		if err != nil {
			return "", err
		}
		return string(sessionID), nil
	}

	sessionID := make([]byte, 32)
	if _, err := rand.Read(sessionID); err != nil {
		return "", err
	}

	sessionID = []byte(hex.EncodeToString(sessionID))
	if err := ioutil.WriteFile(sessionFile, sessionID, 0600); err != nil {
		return "", err
	}

	return string(sessionID), nil
}
