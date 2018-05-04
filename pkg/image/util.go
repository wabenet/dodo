package image

import (
	"encoding/json"
	"errors"
	"io"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/pkg/jsonmessage"
	log "github.com/sirupsen/logrus"
)

var (
	errMissingImageID = errors.New(
		"build complete, but the server did not send an image id")
)

func handleImageResult(result io.ReadCloser, getImageID bool) (string, error) {
	defer func() {
		if err := result.Close(); err != nil {
			log.Error(err)
		}
	}()

	decoder := json.NewDecoder(result)
	imageID := ""

	for {
		var msg jsonmessage.JSONMessage
		if err := decoder.Decode(&msg); err != nil {
			if err == io.EOF {
				break
			}
			return "", err
		}

		if msg.Error != nil {
			return "", msg.Error
		}

		if msg.Progress != nil || msg.ProgressMessage != "" {
			continue
		}

		if msg.Aux != nil && getImageID {
			var result types.BuildResult
			if err := json.Unmarshal(*msg.Aux, &result); err == nil {
				imageID = result.ID
			} else {
				log.Error(err)
			}
			continue
		}

		if msg.Stream != "" {
			log.Debug(strings.TrimRight(msg.Stream, "\n"))
		} else if msg.Status != "" {
			log.Debug(strings.TrimRight(msg.Status, "\n"))
		}
	}

	if getImageID && imageID == "" {
		return "", errMissingImageID
	}

	return imageID, nil
}
