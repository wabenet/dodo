package template

import (
	"bytes"
	"os"
	"os/exec"
	"os/user"
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

func FuncMap() template.FuncMap {
	return template.FuncMap{
		"user": user.Current,
		"cwd":  os.Getwd,
		"env":  os.Getenv,
		"sh":   runShell,
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
