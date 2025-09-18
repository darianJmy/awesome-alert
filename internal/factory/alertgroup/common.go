package alertgroup

import (
	"awesome-alert/internal/ent"
	"awesome-alert/internal/models"
)

func dbToModel(db *ent.AlertGroup) *models.AlertGroup {
	resp := &models.AlertGroup{
		Name:        db.Name,
		Description: db.Description,
		CreatedAt:   db.CreatedAt,
		UpdatedAt:   db.UpdatedAt,
	}

	return resp
}
