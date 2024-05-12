package configbase

import (
	"github.com/lucky-xin/nebula-importer/pkg/client"
	"github.com/lucky-xin/nebula-importer/pkg/logger"
	"github.com/lucky-xin/nebula-importer/pkg/manager"
)

type Configurator interface {
	Optimize(configPath string) error
	Build() error
	GetLogger() logger.Logger
	GetClientPool() client.Pool
	GetManager() manager.Manager
}
