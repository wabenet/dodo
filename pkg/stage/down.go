package stage

import (
	"os"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

func (stage *Stage) Down(remove bool, force bool) error {
	if remove {
		exist, err := stage.provider.Exist()
		if err != nil && !force {
			return err
		}

		if !exist && !force {
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

		available, err := stage.provider.Available()
		if err != nil {
			log.WithFields(log.Fields{"error": err}).Debug("could not get stage status")
		}
		if !available {
			log.WithFields(log.Fields{"name": stage.name}).Info("stage is already down")
			return nil
		}

		if err := stage.provider.Stop(); err != nil {
			return errors.Wrap(err, "could not pause stage")
		}

		for attempts := 0; ; attempts++ {
			available, err := stage.provider.Available()
			if err != nil {
				log.WithFields(log.Fields{"error": err}).Debug("could not get stage status")
			}
			if !available {
				break
			}
			if attempts >= 60 {
				return errors.New("stage did pause successfully")
			}
			time.Sleep(3 * time.Second)
		}

		log.WithFields(log.Fields{"name": stage.name}).Info("paused stage")
	}

	return nil
}
