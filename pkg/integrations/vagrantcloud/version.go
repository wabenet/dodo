package vagrantcloud

import (
	"encoding/json"
	"fmt"
	"net/url"
	"time"
)

type Status string

const (
	Unreleased Status = "unreleased"
	Active     Status = "active"
	Revoked    Status = "revoked"
)

type Version struct {
	Version             string     `json:"version"`
	Status              Status     `json:"status"`
	DescriptionHtml     string     `json:"description_html"`
	DescriptionMarkdown string     `json:"description_markdown"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
	Number              string     `json:"number"`
	Downloads           int        `json:"downloads"`
	ReleaseUrl          string     `json:"release_url"`
	RevokeUrl           string     `json:"revoke_url"`
	Providers           []Provider `json:"providers"`
}

type VersionOptions struct {
	Box         *BoxOptions
	Number      string
	Version     string
	Description string
}

func (v *VersionOptions) toPath() string {
	return fmt.Sprintf("%s/version/%s", v.Box.toPath(), v.Number)
}

func (v *VersionOptions) toParams() url.Values {
	params := url.Values{}
	params.Add("version[version]", v.Version)
	params.Add("version[description]", v.Description)
	return params
}

func (v *VagrantCloud) GetVersion(opts *VersionOptions) (*Version, error) {
	body, err := v.get(opts.toPath())
	if err != nil {
		return nil, err
	}
	version := &Version{}
	if err = json.Unmarshal(body, version); err != nil {
		return nil, err
	}
	return version, nil
}

func (v *VagrantCloud) CreateVersion(opts *VersionOptions) (*Version, error) {
	body, err := v.post(opts.toPath()+"/versions", opts.toParams())
	if err != nil {
		return nil, err
	}
	version := &Version{}
	if err = json.Unmarshal(body, version); err != nil {
		return nil, err
	}
	return version, nil
}

func (v *VagrantCloud) UpdateVersion(opts *VersionOptions) (*Version, error) {
	body, err := v.put(opts.toPath(), opts.toParams())
	if err != nil {
		return nil, err
	}
	version := &Version{}
	if err = json.Unmarshal(body, version); err != nil {
		return nil, err
	}
	return version, nil
}

func (v *VagrantCloud) DeleteVersion(opts *VersionOptions) (*Version, error) {
	body, err := v.delete(opts.toPath())
	if err != nil {
		return nil, err
	}
	version := &Version{}
	if err = json.Unmarshal(body, version); err != nil {
		return nil, err
	}
	return version, nil
}

func (v *VagrantCloud) Release(opts *VersionOptions) (*Version, error) {
	body, err := v.put(opts.toPath()+"/release", url.Values{})
	if err != nil {
		return nil, err
	}
	version := &Version{}
	if err = json.Unmarshal(body, version); err != nil {
		return nil, err
	}
	return version, nil
}

func (v *VagrantCloud) Revoke(opts *VersionOptions) (*Version, error) {
	body, err := v.put(opts.toPath()+"/revoke", url.Values{})
	if err != nil {
		return nil, err
	}
	version := &Version{}
	if err = json.Unmarshal(body, version); err != nil {
		return nil, err
	}
	return version, nil
}
