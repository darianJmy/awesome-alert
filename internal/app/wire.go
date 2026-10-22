//go:build wireinject

package app

import (
	"guanlan/internal/config"
	"io/fs"

	"github.com/google/wire"
)

// InitializeApp initializes configuration, logging, database and HTTP routes.
func InitializeApp(assets fs.FS) (*App, func(), error) {
	wire.Build(config.InitConfig, newLogger, newDatabase, newRouter, newApplication)
	return nil, nil, nil
}
