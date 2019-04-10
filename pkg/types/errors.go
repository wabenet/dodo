package types

import (
	"fmt"
	"reflect"
)

func ErrorUnsupportedType(name string, kind reflect.Kind) error {
	return fmt.Errorf("Unsupported type of '%s': '%v'", name, kind)
}

func ErrorUnsupportedKey(parent string, name string) error {
	return fmt.Errorf("Unsupported option in '%s': '%s'", parent, name)
}
