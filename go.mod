module github.com/dodo-cli/dodo

go 1.15

// TODO: This is currently necessary because of changes in buildkit.
// This part should probably be handled by the code generator.
replace (
	github.com/hashicorp/go-immutable-radix => github.com/tonistiigi/go-immutable-radix v0.0.0-20170803185627-826af9ccf0fe
	github.com/jaguilar/vt100 => github.com/tonistiigi/vt100 v0.0.0-20190402012908-ad4c4a574305
)
