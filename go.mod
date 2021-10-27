module github.com/dodo-cli/dodo

go 1.16

// TODO: This is currently necessary because of changes in buildkit.
// This part should probably be handled by the code generator.
replace (
	github.com/hashicorp/go-immutable-radix => github.com/tonistiigi/go-immutable-radix v0.0.0-20170803185627-826af9ccf0fe
	github.com/jaguilar/vt100 => github.com/tonistiigi/vt100 v0.0.0-20190402012908-ad4c4a574305
)

require (
	github.com/dave/jennifer v1.4.1
	github.com/dodo-cli/dodo-buildkit v0.1.1-0.20211025161657-4e4656afac54
	github.com/dodo-cli/dodo-config v0.1.1-0.20211025161717-2bdef0af0c4c
	github.com/dodo-cli/dodo-core v0.2.0
	github.com/dodo-cli/dodo-docker v0.1.1-0.20211025161757-08f2ee8fc673
	github.com/hashicorp/go-hclog v0.16.2 // indirect
	gopkg.in/yaml.v2 v2.4.0
)
