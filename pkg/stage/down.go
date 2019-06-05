package stage

import (
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

func (stage *Stage) Down(remove bool) error {
	// TODO: support cleanup in case something goes wrong

	if !stage.exists {
		log.WithFields(log.Fields{"name": stage.name}).Info("stage is not up")
		return nil
	}

	if err := stage.host.Stop(); err != nil {
		return errors.Wrap(err, "could not pause stage")
	}
	if err := stage.api.Save(stage.host); err != nil {
		return errors.Wrap(err, "could not store stage")
	}

	if !remove {
		return nil
	}

	if err := stage.host.Driver.Remove(); err != nil {
		return errors.Wrap(err, "could not remove remote stage")
	}

	if err := stage.api.Remove(stage.name); err != nil {
		return errors.Wrap(err, "could not remove local stage")
	}

	return nil
}
