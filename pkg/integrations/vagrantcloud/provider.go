package vagrantcloud

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"time"
)

type Provider struct {
	Name        string    `json:"name"`
	Hosted      bool      `json:"hosted"`
	HostedToken string    `json:"hosted_token"`
	OriginalUrl string    `json:"original_url"`
	UploadUrl   string    `json:"upload_url"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	DownloadUrl string    `json:"download_url"`
}

type ProviderOptions struct {
	Version *VersionOptions
	Name    string
	Url     string
}

func (p *ProviderOptions) toPath() string {
	return fmt.Sprintf("%s/provider/%s", p.Version.toPath(), p.Name)
}

func (p *ProviderOptions) toParams() url.Values {
	params := url.Values{}
	params.Add("provider[name]", p.Name)
	params.Add("provider[url]", p.Url)
	return params
}

func (v *VagrantCloud) GetProvider(opts *ProviderOptions) (*Provider, error) {
	body, err := v.get(opts.toPath())
	if err != nil {
		return nil, err
	}
	provider := &Provider{}
	if err = json.Unmarshal(body, provider); err != nil {
		return nil, err
	}
	return provider, nil
}

func (v *VagrantCloud) CreateProvider(opts *ProviderOptions) (*Provider, error) {
	body, err := v.post(opts.Version.toPath()+"/providers", opts.toParams())
	if err != nil {
		return nil, err
	}
	provider := &Provider{}
	if err = json.Unmarshal(body, provider); err != nil {
		return nil, err
	}
	return provider, nil
}

func (v *VagrantCloud) UpdateProvider(opts *ProviderOptions) (*Provider, error) {
	body, err := v.put(opts.toPath(), opts.toParams())
	if err != nil {
		return nil, err
	}
	provider := &Provider{}
	if err = json.Unmarshal(body, provider); err != nil {
		return nil, err
	}
	return provider, nil
}

func (v *VagrantCloud) DeleteProvider(opts *ProviderOptions) (*Provider, error) {
	body, err := v.delete(opts.toPath())
	if err != nil {
		return nil, err
	}
	provider := &Provider{}
	if err = json.Unmarshal(body, provider); err != nil {
		return nil, err
	}
	return provider, nil
}

func (v *VagrantCloud) UploadProvider(opts *ProviderOptions, data io.Reader) (*Provider, error) {
	body, err := v.upload(opts.toPath()+"/upload", data)
	if err != nil {
		return nil, err
	}
	provider := &Provider{}
	if err = json.Unmarshal(body, provider); err != nil {
		return nil, err
	}
	return provider, nil
}
