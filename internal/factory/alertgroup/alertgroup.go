package alertgroup

import (
	"awesome-alert/internal/ctx"
	"awesome-alert/internal/ent/alertgroup"
	"awesome-alert/internal/models"
	"context"
)

type Interface interface {
	Create(ctx context.Context, req *models.AlertGroup) (interface{}, error)
	Update(ctx context.Context, req *models.AlertGroup) (interface{}, error)
	Delete(ctx context.Context, name string) (interface{}, error)
	Get(ctx context.Context, name string) (*models.AlertGroup, error)
	List(ctx context.Context) ([]*models.AlertGroup, error)
}

type AlertGroup struct {
	ctx *ctx.Context
}

func (a *AlertGroup) Create(ctx context.Context, req *models.AlertGroup) (interface{}, error) {
	return a.ctx.DB.AlertGroup.Create().
		SetName(req.Name).
		SetDescription(req.Description).
		Save(ctx)
}

func (a *AlertGroup) Update(ctx context.Context, req *models.AlertGroup) (interface{}, error) {
	return a.ctx.DB.AlertGroup.Update().
		Where(alertgroup.Name(req.Name)).
		SetDescription(req.Description).
		Save(ctx)
}

func (a *AlertGroup) Delete(ctx context.Context, name string) (interface{}, error) {
	return a.ctx.DB.AlertGroup.Delete().
		Where(alertgroup.Name(name)).
		Exec(ctx)
}

func (a *AlertGroup) Get(ctx context.Context, name string) (*models.AlertGroup, error) {
	db, err := a.ctx.DB.AlertGroup.Query().
		Where(alertgroup.Name(name)).
		Only(ctx)
	if err != nil {
		return nil, err
	}

	return dbToModel(db), nil
}

func (a *AlertGroup) List(ctx context.Context) ([]*models.AlertGroup, error) {
	dbs, err := a.ctx.DB.AlertGroup.Query().
		All(ctx)
	if err != nil {
		return nil, err
	}

	list := make([]*models.AlertGroup, 0, len(dbs))
	for _, db := range dbs {
		list = append(list, dbToModel(db))
	}

	return list, nil
}

func NewAlertGroup(ctx *ctx.Context) *AlertGroup {
	return &AlertGroup{
		ctx: ctx,
	}
}
