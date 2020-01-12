package types

import (
	"bytes"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"reflect"
	"text/template"

	"github.com/Masterminds/sprig"
)

type decoder struct {
	filename string
}

func NewDecoder(filename string) *decoder {
	return &decoder{filename: filename}
}

func (d *decoder) WithFile(filename string) *decoder {
	return &decoder{filename: filename}
}

func (d *decoder) DecodeBool(name string, config interface{}) (bool, error) {
	var result bool
	switch t := reflect.ValueOf(config); t.Kind() {
	case reflect.Bool:
		result = t.Bool()
	default:
		return result, &ConfigError{Name: name, UnsupportedType: t.Kind()}
	}
	return result, nil
}

func (d *decoder) DecodeInt(name string, config interface{}) (int64, error) {
	var result int64
	switch t := reflect.ValueOf(config); t.Kind() {
	case reflect.Int:
		result = t.Int()
	default:
		return result, &ConfigError{Name: name, UnsupportedType: t.Kind()}
	}
	return result, nil
}

func (d *decoder) DecodeString(name string, config interface{}) (string, error) {
	var result string
	switch t := reflect.ValueOf(config); t.Kind() {
	case reflect.String:
		return d.ApplyTemplate(t.String())
	default:
		return result, &ConfigError{Name: name, UnsupportedType: t.Kind()}
	}
}

func (d *decoder) DecodeStringSlice(name string, config interface{}) ([]string, error) {
	var result []string
	switch t := reflect.ValueOf(config); t.Kind() {
	case reflect.String:
		decoded, err := d.DecodeString(name, t.String())
		if err != nil {
			return result, err
		}
		result = []string{decoded}
	case reflect.Slice:
		for _, v := range t.Interface().([]interface{}) {
			decoded, err := d.DecodeString(name, v)
			if err != nil {
				return result, err
			}
			result = append(result, decoded)
		}
	default:
		return result, &ConfigError{Name: name, UnsupportedType: t.Kind()}
	}
	return result, nil
}

func (d *decoder) ApplyTemplate(input string) (string, error) {
	templ, err := template.New("config").Funcs(sprig.TxtFuncMap()).Funcs(d.FuncMap()).Parse(input)
	if err != nil {
		return "", err
	}

	var buffer bytes.Buffer
	err = templ.Execute(&buffer, *d)
	if err != nil {
		return "", err
	}

	return buffer.String(), nil
}

func (d *decoder) FuncMap() template.FuncMap {
	return template.FuncMap{
		"user": user.Current,
		"cwd":  os.Getwd,
		"env":  os.Getenv,
		"sh":   runShell,
		"projectRoot": func() (string, error) {
			root, _, err := findProjectRoot()
			return root, err
		},
		"projectPath": func() (string, error) {
			_, path, err := findProjectRoot()
			return path, err
		},
		"currentFile": func() string {
			return d.filename
		},
		"currentDir": func() string {
			return filepath.Dir(d.filename)
		},
	}
}

func runShell(command string) (string, error) {
	// TODO: what to do on windows?
	cmd := exec.Command("/bin/sh", "-c", command)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return "", err
	}
	return out.String(), nil
}

func findProjectRoot() (string, string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", "", err
	}
	for dir := cwd; dir != "/"; dir = filepath.Dir(dir) {
		if info, err := os.Stat(filepath.Join(dir, ".git")); err == nil && info.IsDir() {
			path, err2 := filepath.Rel(dir, cwd)
			if err2 != nil {
				return "", "", err
			}
			return dir, path, nil
		}
	}
	return cwd, ".", nil
}
