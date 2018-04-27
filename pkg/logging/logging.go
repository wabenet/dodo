package logging

import (
	"bytes"
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
)

// Options represent the logging options
type Options struct {
	Quiet bool
	Debug bool
}

type formatter struct{}

// Format implements the log.Formatter interface
func (f *formatter) Format(entry *log.Entry) ([]byte, error) {
	b := &bytes.Buffer{}
	fmt.Fprintf(b, "%-44s ", entry.Message)
	for key, value := range entry.Data {
		b.WriteString(key)
		b.WriteByte('=')
		stringValue, ok := value.(string)
		if !ok {
			stringValue = fmt.Sprint(value)
		}
		b.WriteString(stringValue)
	}
	b.WriteString("\n")
	return b.Bytes(), nil
}

// InitFlags adds logging related flags to a flag set
func InitFlags(flags *pflag.FlagSet, opts *Options) {
	flags.BoolVarP(
		&opts.Quiet, "quiet", "q", false,
		"suppress informational output")
	flags.BoolVarP(
		&opts.Debug, "debug", "", false,
		"show additional debug output")
}

// InitLogging initializes the logging framework based on the passed options
func InitLogging(opts *Options) {
	log.SetFormatter(&formatter{})
	if opts.Quiet {
		log.SetLevel(log.WarnLevel)
	} else if opts.Debug {
		log.SetLevel(log.DebugLevel)
	}
}
