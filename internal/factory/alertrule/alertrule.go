package alertrule

import (
	"awesome-alert/internal/ctx"
	"awesome-alert/internal/ent"
	"awesome-alert/internal/ent/alertrule"
	"awesome-alert/internal/ent/datasource"
	"awesome-alert/internal/models"
	"awesome-alert/internal/types"
	"context"
)

type Interface interface {
	CreateAlertRule(ctx context.Context, req *models.AlertRule) (interface{}, error)
	UpdateAlertRule(ctx context.Context, req *models.AlertRule) (interface{}, error)
	DeleteAlertRule(ctx context.Context, name string) (interface{}, error)
	GetAlertRule(ctx context.Context, name string) (*models.AlertRule, error)
	ListAlertRule(ctx context.Context, req *models.RequestAlertRuleList) (interface{}, error)
}

type AlertRule struct {
	ctx *ctx.Context
}

func (a *AlertRule) CreateAlertRule(ctx context.Context, req *models.AlertRule) (interface{}, error) {
	return a.ctx.DB.AlertRule.Create().
		SetGroup(req.Group).
		SetName(req.Name).
		SetAddLabel(req.AdditionalLabel).
		SetType(req.Type).
		SetDatasourceID(req.DataSourceId).
		SetExpr(req.Expr).
		SetCondition(req.Condition).
		SetDuration(req.Duration).
		SetAlertContinuousCount(req.AlertContinuousCount).
		SetRecoverContinuousCount(req.RecoverContinuousCount).
		SetSendRecoverEvent(req.SendRecoverEvent).
		SetFaultCenter(req.FaultCenterId).
		SetEnabled(req.Enabled).
		SetCreatedAt(req.CreatedAt).
		SetUpdatedAt(req.UpdatedAt).
		Save(ctx)
}

func (a *AlertRule) UpdateAlertRule(ctx context.Context, req *models.AlertRule) (interface{}, error) {
	return a.ctx.DB.AlertRule.Update().Where(alertrule.Name(req.Name)).
		SetGroup(req.Group).
		SetAddLabel(req.AdditionalLabel).
		SetType(req.Type).
		SetDatasourceID(req.DataSourceId).
		SetExpr(req.Expr).
		SetCondition(req.Condition).
		SetDuration(req.Duration).
		SetAlertContinuousCount(req.AlertContinuousCount).
		SetRecoverContinuousCount(req.RecoverContinuousCount).
		SetSendRecoverEvent(req.SendRecoverEvent).
		SetFaultCenter(req.FaultCenterId).
		SetEnabled(req.Enabled).
		Save(ctx)
}

func (a *AlertRule) DeleteAlertRule(ctx context.Context, name string) (interface{}, error) {
	return a.ctx.DB.AlertRule.Delete().Where(alertrule.Name(name)).Exec(ctx)
}

func (a *AlertRule) GetAlertRule(ctx context.Context, name string) (*models.AlertRule, error) {
	db, err := a.ctx.DB.AlertRule.Query().
		Where(alertrule.Name(name)).
		Only(ctx)
	if err != nil {
		return nil, err
	}

	return dbToModel(db), nil
}

func (a *AlertRule) ListAlertRule(ctx context.Context, req *models.RequestAlertRuleList) (interface{}, error) {
	if req.Page <= 0 || req.PageSize <= 0 {
		req.Page, req.PageSize = 1, 20
	}

	db := a.ctx.DB.AlertRule.Query()
	total, err := db.Count(ctx)
	if err != nil {
		return nil, err
	}

	if req.Group != "" {
		db = db.Where(alertrule.GroupContains(req.Group))
	}
	if req.Name != "" {
		db = db.Where(alertrule.NameContains(req.Name))
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

	list := make([]*models.AlertRule, 0, len(dbs))
	for _, v := range dbs {
		list = append(list, dbToModel(v))
	}

	return &models.AlertRuleList{
		List:       list,
		Pagination: pagination,
	}, nil
}

func NewAlertRule(ctx *ctx.Context) *AlertRule {
	return &AlertRule{
		ctx: ctx,
	}
}
