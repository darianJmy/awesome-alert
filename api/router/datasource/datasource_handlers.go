package datasource

import (
	"awesome-alert/internal/factory"
	"awesome-alert/internal/models"
	"awesome-alert/pkg/httputils"
	"context"
	"github.com/cloudwego/hertz/pkg/app"
)

func (d *dataSourceRouter) createDataSource(ctx context.Context, c *app.RequestContext) {
	var req models.DataSource
	if err := c.BindJSON(&req); err != nil {
		httputils.Failed(c, err)
		return
	}

	_, err := factory.Factory.DataSource().Create(ctx, &req)
	if err != nil {
		httputils.Failed(c, err)
		return
	}

	httputils.Success(c, nil)
}

func (d *dataSourceRouter) updateDataSource(ctx context.Context, c *app.RequestContext) {
	var req models.DataSource
	if err := c.BindJSON(&req); err != nil {
		httputils.Failed(c, err)
		return
	}

	_, err := factory.Factory.DataSource().Update(ctx, &req)
	if err != nil {
		httputils.Failed(c, err)
		return
	}

	httputils.Success(c, nil)
}

func (d *dataSourceRouter) deleteDataSource(ctx context.Context, c *app.RequestContext) {
	_, err := factory.Factory.DataSource().Delete(ctx, c.Query("name"))
	if err != nil {
		httputils.Failed(c, err)
		return
	}

	httputils.Success(c, nil)
}

func (d *dataSourceRouter) getDataSource(ctx context.Context, c *app.RequestContext) {
	result, err := factory.Factory.DataSource().Get(ctx, c.Query("name"))
	if err != nil {
		httputils.Failed(c, err)
		return
	}

	httputils.Success(c, result)
}

func (d *dataSourceRouter) listDataSource(ctx context.Context, c *app.RequestContext) {
	var req models.RequestDataSourceList
	if err := c.BindQuery(&req); err != nil {
		httputils.Failed(c, err)
		return
	}

	result, err := factory.Factory.DataSource().List(ctx, &req)
	if err != nil {
		httputils.Failed(c, err)
		return
	}

	httputils.Success(c, result)
}
