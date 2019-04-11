package template

import (
	"bytes"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"text/template"

	"github.com/Masterminds/sprig"
)

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

func FuncMap() template.FuncMap {
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
	}
}

func ApplyTemplate(input string) (string, error) {
	templ, err := template.New("config").Funcs(sprig.TxtFuncMap()).Funcs(FuncMap()).Parse(input)
	if err != nil {
		return "", err
	}

	var buffer bytes.Buffer
	var data struct{}
	err = templ.Execute(&buffer, data)
	if err != nil {
		return "", err
	}

	return buffer.String(), nil
}
