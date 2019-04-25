package types

import (
	"fmt"
	"reflect"
)

type ConfigError struct {
	Name            string
	UnsupportedKey  *string
	UnsupportedType reflect.Kind
}

func (e *ConfigError) Error() string {
	if e.UnsupportedKey != nil {
		return fmt.Sprintf("unsupported option in '%s': '%s'", e.Name, *e.UnsupportedKey)
	}
	if e.UnsupportedType != reflect.Invalid {
		return fmt.Sprintf("unsupported type of '%s': '%s'", e.Name, e.UnsupportedType.String())
	}
	return fmt.Sprintf("configuration error in '%s'", e.Name)
}

type ScriptError struct {
	Message  string
	ExitCode int
}

func (e *ScriptError) Error() string {
	return e.Message
}
