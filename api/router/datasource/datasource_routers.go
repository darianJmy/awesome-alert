package datasource

import "github.com/cloudwego/hertz/pkg/app/server"

type dataSourceRouter struct {
}

func NewRouter(hz *server.Hertz) {
	router := &dataSourceRouter{}
	router.initRoute(hz)
}

func (d *dataSourceRouter) initRoute(hz *server.Hertz) {
	authRoute := hz.Group("/v1/datasource")
	{
		authRoute.POST("/create", d.createDataSource)
		authRoute.PUT("/update", d.updateDataSource)
		authRoute.DELETE("/delete", d.deleteDataSource)
		authRoute.GET("/get", d.getDataSource)
		authRoute.GET("/list", d.listDataSource)

	}

}
