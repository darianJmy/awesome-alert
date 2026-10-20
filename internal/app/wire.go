//go:build wireinject

package app

import (
	"guanlan/internal/config"

	"github.com/google/wire"
)

// InitializeApp initializes configuration, logging, database and HTTP routes.
func InitializeApp() (*App, func(), error) {
	wire.Build(config.InitConfig, embeddedAssets, newLogger, newDatabase, newRouter, newApplication)
	return nil, nil, nil
}
