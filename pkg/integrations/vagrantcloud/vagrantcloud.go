package vagrantcloud

import (
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

const (
	baseUrl = "https://vagrantcloud.com/api/v1"
)

type VagrantCloud struct {
	accessToken string
}

func New(accessToken string) *VagrantCloud {
	return &VagrantCloud{accessToken: accessToken}
}

func (v *VagrantCloud) get(path string) ([]byte, error) {
	return v.request(
		"GET",
		path,
		"application/x-www-form-urlencoded",
		strings.NewReader(""),
	)
}

func (v *VagrantCloud) post(path string, params url.Values) ([]byte, error) {
	return v.request(
		"POST",
		path,
		"application/x-www-form-urlencoded",
		strings.NewReader(params.Encode()),
	)
}

func (v *VagrantCloud) put(path string, params url.Values) ([]byte, error) {
	return v.request(
		"PUT",
		path,
		"application/x-www-form-urlencoded",
		strings.NewReader(params.Encode()),
	)
}

func (v *VagrantCloud) delete(path string) ([]byte, error) {
	return v.request(
		"DELETE",
		path,
		"application/x-www-form-urlencoded",
		strings.NewReader(""),
	)
}

func (v *VagrantCloud) upload(path string, data io.Reader) ([]byte, error) {
	return v.request(
		"PUT",
		path,
		"multipart/form-data",
		data,
	)
}

func (v *VagrantCloud) request(method string, path string, contentType string, data io.Reader) ([]byte, error) {
	requestUri, err := url.ParseRequestURI(baseUrl + path)
	if err != nil {
		return nil, err
	}

	if v.accessToken != "" {
		query := requestUri.Query()
		query.Set("access_token", v.accessToken)
		requestUri.RawQuery = query.Encode()
	}

	req, err := http.NewRequest(method, requestUri.String(), data)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, errors.New(string(resp.StatusCode))
	}
	return body, nil
}
