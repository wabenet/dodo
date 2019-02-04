package types

import (
	"fmt"
)

// KeyValueList represents a list of key/value pairs
type KeyValueList []KeyValue

// KeyValue represents a key/value pair, where the value is optional
type KeyValue struct {
	Key   string
	Value *string
}

// Strings transforms a key/value list into a list of strings that will be
// understood by docker.
func (kvs *KeyValueList) Strings() []string {
	result := []string{}
	for _, kv := range *kvs {
		result = append(result, kv.String())
	}
	return result
}

func (kv *KeyValue) String() string {
	if kv.Value == nil {
		return kv.Key
	}
	return fmt.Sprintf("%s=%s", kv.Key, *kv.Value)
}
