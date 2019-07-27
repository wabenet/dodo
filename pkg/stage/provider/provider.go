package provider

var BuiltInProviders = map[string]Provider{
	DefaultProviderName: &DefaultProvider{},
}

type SSHOptions struct {
	Hostname string
	Port     int
	Username string
}

type DockerOptions struct {
	Version  string
	Host     string
	CAFile   string
	CertFile string
	KeyFile  string
}

type Provider interface {
	Initialize(map[string]string) (bool, error)
	Create() error
	Start() error
	Stop() error
	Remove(bool) error
	Exist() (bool, error)
	Available() (bool, error)
	GetURL() (string, error)
	GetIP() (string, error)
	GetSSHOptions() (*SSHOptions, error)
	GetDockerOptions() (*DockerOptions, error)
}
