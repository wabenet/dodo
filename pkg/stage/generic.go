package stage

type GenericStage struct {
	Options *DockerOptions
}

func (stage *GenericStage) Initialize(_ string, opts map[string]string) (bool, error) {
	stage.Options = &DockerOptions{
		Version:  opts["api_version"],
		Host:     opts["host"],
		CAFile:   opts["ca_file"],
		CertFile: opts["cert_file"],
		KeyFile:  opts["key_file"],
	}
	return true, nil
}

func (stage *GenericStage) Create() error {
	return nil
}

func (stage *GenericStage) Start() error {
	return nil
}

func (stage *GenericStage) Stop() error {
	return nil
}

func (stage *GenericStage) Remove(_ bool) error {
	return nil
}

func (stage *GenericStage) Exist() (bool, error) {
	return true, nil
}

func (stage *GenericStage) Available() (bool, error) {
	return true, nil
}

func (stage *GenericStage) GetSSHOptions() (*SSHOptions, error) {
	return nil, nil
}

func (stage *GenericStage) GetDockerOptions() (*DockerOptions, error) {
	return stage.Options, nil
}
