package stage

import (
	"github.com/docker/machine/libmachine/state"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

func (stage *Stage) Down(remove bool, force bool) error {
	if remove {
		if !stage.exists && !force {
			log.WithFields(log.Fields{"name": stage.name}).Info("stage is not up")
			return nil
		}

		if err := stage.driver.Remove(); err != nil && !force {
			return errors.Wrap(err, "could not remove remote stage")
		}

		if err := stage.deleteState(); err != nil && !force {
			return errors.Wrap(err, "could not remove local stage")
		}

		log.WithFields(log.Fields{"name": stage.name}).Info("removed stage")
	} else {
		log.WithFields(log.Fields{"name": stage.name}).Info("pausing stage...")

		currentState, err := stage.driver.GetState()
		if err != nil {
			log.WithFields(log.Fields{"error": err}).Debug("could not get machine state")
		}
		if currentState == state.Stopped {
			log.WithFields(log.Fields{"name": stage.name}).Info("stage is already down")
			return nil
		}

		if err := stage.driver.Stop(); err != nil {
			return errors.Wrap(err, "could not pause stage")
		}

		if err := stage.waitForState(state.Stopped); err != nil {
			return errors.Wrap(err, "could not pause stage")
		}

		if err := stage.exportState(); err != nil && !force {
			return errors.Wrap(err, "could not store stage")
		}

		log.WithFields(log.Fields{"name": stage.name}).Info("paused stage")
	}

	return nil
}
