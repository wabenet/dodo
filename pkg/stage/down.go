package stage

import (
	"os"

	vbox "github.com/oclaussen/dodo/pkg/stage/virtualbox"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

func (stage *Stage) Down(remove bool, force bool) error {
	if remove {
		if !stage.exists && !force {
			log.WithFields(log.Fields{"name": stage.name}).Info("stage is not up")
			return nil
		}

		if err := vbox.Remove(stage.name); err != nil && !force {
			return errors.Wrap(err, "could not remove remote stage")
		}

		if err := os.RemoveAll(stage.hostDir()); err != nil && !force {
			return errors.Wrap(err, "could not remove local stage")
		}

		log.WithFields(log.Fields{"name": stage.name}).Info("removed stage")
	} else {
		log.WithFields(log.Fields{"name": stage.name}).Info("pausing stage...")

		currentStatus, err := vbox.GetStatus(stage.name)
		if err != nil {
			log.WithFields(log.Fields{"error": err}).Debug("could not get machine status")
		}
		if currentStatus == vbox.Stopped {
			log.WithFields(log.Fields{"name": stage.name}).Info("stage is already down")
			return nil
		}

		if err := vbox.Stop(stage.name); err != nil {
			return errors.Wrap(err, "could not pause stage")
		}

		if err := stage.waitForStatus(vbox.Stopped); err != nil {
			return errors.Wrap(err, "could not pause stage")
		}

		log.WithFields(log.Fields{"name": stage.name}).Info("paused stage")
	}

	return nil
}
