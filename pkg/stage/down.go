package stage

import (
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

func (stage *Stage) Down(remove bool, force bool) error {
	if remove {
		if !stage.exists && !force {
			log.WithFields(log.Fields{"name": stage.name}).Info("stage is not up")
			return nil
		}

		if err := stage.host.Driver.Remove(); err != nil && !force {
			return errors.Wrap(err, "could not remove remote stage")
		}

		if err := stage.api.Remove(stage.name); err != nil && !force {
			return errors.Wrap(err, "could not remove local stage")
		}

		log.WithFields(log.Fields{"name": stage.name}).Info("removed stage")
	} else {
		if err := stage.host.Stop(); err != nil && !force {
			return errors.Wrap(err, "could not pause stage")
		}
		if err := stage.api.Save(stage.host); err != nil && !force {
			return errors.Wrap(err, "could not store stage")
		}

		log.WithFields(log.Fields{"name": stage.name}).Info("paused stage")
	}

	return nil
}
