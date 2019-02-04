package types

import (
	"fmt"
)

// Volumes represents a set of volume configurations
type Volumes []Volume

// Volume represents a bind-mount volume configuration
type Volume struct {
	Source   string
	Target   string
	ReadOnly bool
}

// Strings transforms a set of Volume definitions into a list of strings that
// will be understood by docker.
func (vs *Volumes) Strings() []string {
	result := []string{}
	for _, v := range *vs {
		result = append(result, v.String())
	}
	return result
}

func (v *Volume) String() string {
	if v.Target == "" && !v.ReadOnly {
		return fmt.Sprintf("%s:%s", v.Source, v.Source)
	} else if !v.ReadOnly {
		return fmt.Sprintf("%s:%s", v.Source, v.Target)
	} else {
		return fmt.Sprintf("%s:%s:ro", v.Source, v.Target)
	}
}
