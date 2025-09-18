package router

import (
	"awesome-alert/api/router/datasource"
	"github.com/cloudwego/hertz/pkg/app/server"
)

type registerFunc func(hz *server.Hertz)

func InstallRouters(hz *server.Hertz) {
	fs := []registerFunc{
		datasource.NewRouter,
	}

	install(hz, fs...)
}

func install(hz *server.Hertz, fs ...registerFunc) {
	for _, f := range fs {
		f(hz)
	}
}
