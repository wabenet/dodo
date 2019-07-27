package provider

type Status int

const (
	Unknown Status = iota
	Down
	Up
	Paused
	Error
)

type Options struct {
	CPU      int
	Memory   int
	DiskSize int
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
	Status() (Status, error)
	Create() error
	Start() error
	Stop() error
	Remove() error
	GetURL() (string, error)
	GetIP() (string, error)
	GetSSHOptions() (*SSHOptions, error)
	GetDockerOptions() (*DockerOptions, error)
}

var BuiltInProviders = map[string]Provider{
	DefaultProviderName: &DefaultProvider{},
}
