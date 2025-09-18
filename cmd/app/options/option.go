package options

import (
	"awesome-alert/cmd/app/config"
	"awesome-alert/internal/ent"
	"context"
	"fmt"
	"github.com/cloudwego/hertz/pkg/app/server"
	_ "github.com/go-sql-driver/mysql"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
)

const (
	configFile = "config/config.yaml"
)

type Options struct {
	ctx context.Context

	ConfigFile string
	Config     config.Config

	EntClient   *ent.Client
	HertzServer *server.Hertz
}

func NewOptions() *Options {
	return &Options{
		ctx: context.Background(),
	}
}

func (o *Options) Complete() error {
	if len(o.ConfigFile) == 0 {
		if cfgFile := os.Getenv("ConfigFile"); cfgFile != "" {
			o.ConfigFile = cfgFile
		} else {
			o.ConfigFile = configFile
		}
	}

	data, err := ioutil.ReadFile(o.ConfigFile)
	if err != nil {
		return err
	}

	var cfg config.Config
	if err = yaml.Unmarshal(data, &cfg); err != nil {
		return err
	}

	o.Config = cfg

	return nil
}

func (o *Options) Register() error {
	if err := o.registerEntClient(); err != nil {
		return err
	}

	o.registerHertzServer()

	return nil
}

func (o *Options) registerEntClient() error {
	sql := o.Config.MYSQL
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		sql.User,
		sql.Password,
		sql.Host,
		sql.Port,
		sql.Database)

	entClient, err := ent.Open("mysql", dsn)
	if err != nil {
		return err
	}

	if err = entClient.Schema.Create(o.ctx); err != nil {
		return err
	}

	o.EntClient = entClient

	return nil
}

func (o *Options) registerHertzServer() {
	listen := o.Config.Default.Listen
	o.HertzServer = server.Default(
		server.WithHostPorts(fmt.Sprintf(":%d", listen)),
		server.WithMaxRequestBodySize(10*1024*1024),
	)
}

func (o *Options) BindFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&o.ConfigFile, "file", o.ConfigFile, "The location of the app configuration file")
}
