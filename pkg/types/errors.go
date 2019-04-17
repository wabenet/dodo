package types

import (
	"fmt"
	"reflect"
)

func ErrorUnsupportedType(name string, kind reflect.Kind) error {
	return fmt.Errorf("unsupported type of '%s': '%v'", name, kind)
}

func ErrorUnsupportedKey(parent string, name string) error {
	return fmt.Errorf("unsupported option in '%s': '%s'", parent, name)
}

type ScriptError struct {
	Message  string
	ExitCode int
}

func (e *ScriptError) Error() string {
	return e.Message
}
