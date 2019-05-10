package configfiles

import (
	"errors"
	"path/filepath"
	"strings"
)

type Options struct {
	Name                      string
	Extensions                []string
	FileGlobs                 []string
	UseFileGlobsOnly          bool
	IncludeWorkingDirectories bool
	Filter                    func(*ConfigFile) bool
}

func (opts *Options) normalize() error {
	if len(opts.Name) == 0 && len(opts.FileGlobs) == 0 {
		return errors.New("either Name or FileGlobs are required")
	}

	if len(opts.Name) > 0 && len(opts.Extensions) == 0 {
		if ext := filepath.Ext(opts.Name); len(ext) > 0 {
			opts.Extensions = []string{ext}
			opts.Name = strings.TrimSuffix(opts.Name, ext)
		}
	}

	if len(opts.FileGlobs) == 0 {
		// TODO: print warning
		opts.UseFileGlobsOnly = false
	}

	if opts.Filter == nil {
		opts.Filter = func(_ *ConfigFile) bool { return true }
	}

	return nil
}
