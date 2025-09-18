package alertrule

import (
	"awesome-alert/internal/ent"
	"awesome-alert/internal/models"
)

func dbToModel(db *ent.AlertRule) *models.AlertRule {
	resp := &models.AlertRule{
		Group:                  db.Group,
		Name:                   db.Name,
		AdditionalLabel:        db.AddLabel,
		Type:                   db.Type,
		DataSourceId:           db.DatasourceID,
		Expr:                   db.Expr,
		Condition:              db.Condition,
		Duration:               db.Duration,
		AlertContinuousCount:   db.AlertContinuousCount,
		RecoverContinuousCount: db.RecoverContinuousCount,
		SendRecoverEvent:       db.SendRecoverEvent,
		FaultCenterId:          db.FaultCenter,
		Enabled:                db.Enabled,
		CreatedAt:              db.CreatedAt,
		UpdatedAt:              db.UpdatedAt,
	}

	return resp
}
