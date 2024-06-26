package configbase

import (
	"path/filepath"

	"github.com/lucky-xin/nebula-importer/pkg/logger"
	"github.com/lucky-xin/nebula-importer/pkg/utils"
)

type Log struct {
	Level   *string       `yaml:"level,omitempty" json:"level,omitempty,optional"`
	Console *bool         `yaml:"console,omitempty" json:"console,omitempty,optional"`
	Files   []string      `yaml:"files,omitempty" json:"files,omitempty,optional"`
	Fields  logger.Fields `yaml:"fields,omitempty" json:"fields,omitempty,optional"`
}

// OptimizePath optimizes relative paths base to the configuration file path
func (l *Log) OptimizePath(configPath string) error {
	if l == nil {
		return nil
	}

	configPathDir := filepath.Dir(configPath)
	for i := range l.Files {
		l.Files[i] = utils.RelativePathBaseOn(configPathDir, l.Files[i])
	}

	return nil
}

func (l *Log) BuildLogger(opts ...logger.Option) (logger.Logger, error) {
	options := make([]logger.Option, 0, 4+len(opts))
	if l != nil {
		if l.Level != nil && *l.Level != "" {
			options = append(options, logger.WithLevelText(*l.Level))
		}
		if l.Console != nil {
			options = append(options, logger.WithConsole(*l.Console))
		}
		if len(l.Files) > 0 {
			options = append(options, logger.WithFiles(l.Files...))
		}
		if len(l.Fields) > 0 {
			options = append(options, logger.WithFields(l.Fields...))
		}
	}
	options = append(options, opts...)
	return logger.New(options...)
}
