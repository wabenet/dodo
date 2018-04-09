package image

import (
	"encoding/json"
	"errors"
	"io"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/docker/docker/pkg/term"
	log "github.com/sirupsen/logrus"
)

var (
	errMissingImageID = errors.New(
		"build complete, but the server did not send an image id",
	)
)

func handleImageResult(result io.ReadCloser, getImageID bool) (string, error) {
	defer func() {
		if err := result.Close(); err != nil {
			log.Error(err)
		}
	}()

	imageID := ""
	aux := func(auxJSON *json.RawMessage) {
		if !getImageID {
			return
		}
		var result types.BuildResult
		if err := json.Unmarshal(*auxJSON, &result); err == nil {
			imageID = result.ID
		} else {
			log.Error(err)
		}
	}

	outFd, isTerminal := term.GetFdInfo(os.Stdout)
	err := jsonmessage.DisplayJSONMessagesStream(
		result,
		os.Stdout,
		outFd,
		isTerminal,
		aux,
	)
	if err != nil {
		return "", err
	}

	if getImageID && imageID == "" {
		return "", errMissingImageID
	}

	return imageID, err
}
