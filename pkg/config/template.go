package config

import (
	"bytes"
	"os"
	"os/exec"
	"os/user"
	"text/template"
)

var functions = template.FuncMap{
	"user": user.Current,
	"cwd":  os.Getwd,
	"env":  os.Getenv,
	// TODO: what to do on windows?
	"sh": func(command string) (string, error) {
		cmd := exec.Command("/bin/sh", "-c", command)
		var out bytes.Buffer
		cmd.Stdout = &out
		err := cmd.Run()
		if err != nil {
			return "", err
		}
		return out.String(), nil
	},
}

func ApplyTemplate(input string) (string, error) {
	templ, err := template.New("config").Funcs(functions).Parse(input)
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
