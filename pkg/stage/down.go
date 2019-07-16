package stage

import (
	"os"

	"github.com/oclaussen/dodo/pkg/stage/provider"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

func (stage *Stage) Down(remove bool, force bool) error {
	if remove {
		if !stage.exists && !force {
			log.WithFields(log.Fields{"name": stage.name}).Info("stage is not up")
			return nil
		}

		if err := stage.provider.Remove(); err != nil && !force {
			return errors.Wrap(err, "could not remove remote stage")
		}

		if err := os.RemoveAll(stage.hostDir()); err != nil && !force {
			return errors.Wrap(err, "could not remove local stage")
		}

		log.WithFields(log.Fields{"name": stage.name}).Info("removed stage")
	} else {
		log.WithFields(log.Fields{"name": stage.name}).Info("pausing stage...")

		currentStatus, err := stage.provider.Status()
		if err != nil {
			log.WithFields(log.Fields{"error": err}).Debug("could not get machine status")
		}
		if currentStatus == provider.Paused {
			log.WithFields(log.Fields{"name": stage.name}).Info("stage is already down")
			return nil
		}

		if err := stage.provider.Stop(); err != nil {
			return errors.Wrap(err, "could not pause stage")
		}

		if err := stage.waitForStatus(provider.Paused); err != nil {
			return errors.Wrap(err, "could not pause stage")
		}

		log.WithFields(log.Fields{"name": stage.name}).Info("paused stage")
	}

	return nil
}
