package datasource

import (
	"awesome-alert/internal/ctx"
	"awesome-alert/internal/ent"
	"awesome-alert/internal/ent/datasource"
	"awesome-alert/internal/models"
	"awesome-alert/internal/types"
	"context"
)

type Interface interface {
	Create(ctx context.Context, req *models.DataSource) (interface{}, error)
	Update(ctx context.Context, req *models.DataSource) (interface{}, error)
	Delete(ctx context.Context, name string) (interface{}, error)
	Get(ctx context.Context, name string) (interface{}, error)
	List(ctx context.Context, req *models.RequestDataSourceList) (interface{}, error)
}

type DataSource struct {
	ctx *ctx.Context
}

func (d *DataSource) Create(ctx context.Context, req *models.DataSource) (interface{}, error) {
	config, err := getConfigMap(req)
	if err != nil {
		return nil, err
	}

	return d.ctx.DB.DataSource.Create().
		SetName(req.Name).
		SetType(req.Type).
		SetDescription(req.Description).
		SetConfig(config).
		Save(ctx)
}

func (d *DataSource) Update(ctx context.Context, req *models.DataSource) (interface{}, error) {
	config, err := getConfigMap(req)
	if err != nil {
		return nil, err
	}

	return d.ctx.DB.DataSource.Update().
		Where(datasource.Name(req.Name)).
		SetDescription(req.Description).
		SetConfig(config).
		Save(ctx)
}

func (d *DataSource) Delete(ctx context.Context, name string) (interface{}, error) {
	return d.ctx.DB.DataSource.Delete().
		Where(datasource.Name(name)).
		Exec(ctx)
}

func (d *DataSource) Get(ctx context.Context, name string) (interface{}, error) {
	db, err := d.ctx.DB.DataSource.Query().
		Where(datasource.Name(name)).
		Only(ctx)
	if err != nil {
		return nil, err
	}

	return dbToModel(db), nil
}

func (d *DataSource) List(ctx context.Context, req *models.RequestDataSourceList) (interface{}, error) {
	if req.Page <= 0 || req.PageSize <= 0 {
		req.Page, req.PageSize = 1, 20
	}

	db := d.ctx.DB.DataSource.Query()
	total, err := db.Count(ctx)
	if err != nil {
		return nil, err
	}

	if req.Name != "" {
		db = db.Where(datasource.NameContains(req.Name))
	}

	pagination := &types.Pagination{
		Page:     req.Page,
		PageSize: req.PageSize,
		Total:    total,
	}

	offset := (req.Page - 1) * req.PageSize

	dbs, err := db.Limit(req.PageSize).Offset(offset).
		Order(ent.Desc(datasource.FieldID)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	list := make([]*models.DataSource, 0, len(dbs))
	for _, v := range dbs {
		list = append(list, dbToModel(v))
	}

	return &models.DataSourceList{
		List:       list,
		Pagination: pagination,
	}, nil
}

func NewDataSource(ctx *ctx.Context) *DataSource {
	return &DataSource{
		ctx: ctx,
	}
}
