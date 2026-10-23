// Package config loads command-line configuration and publishes the application settings.
package config

import (
	"flag"
	"fmt"
	"guanlan/internal/conf"
	"os"
)

type Config = conf.Config

var AppConfig = conf.DefaultConfig()

// InitConfig publishes configuration once, before database and handler initialization.
func InitConfig() (Config, error) {
	flags := flag.NewFlagSet("guanlan", flag.ContinueOnError)
	path := flags.String("conf", "", "configuration YAML path (default: optional ./conf/config.yaml)")
	if err := flags.Parse(os.Args[1:]); err != nil {
		return Config{}, err
	}
	if flags.NArg() != 0 {
		return Config{}, fmt.Errorf("unexpected arguments: %v", flags.Args())
	}
	cfg, err := conf.InitConfig(*path)
	if err != nil {
		return Config{}, err
	}
	AppConfig = cfg
	return cfg, nil
}
