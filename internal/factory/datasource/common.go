package datasource

import (
	"awesome-alert/internal/ent"
	"awesome-alert/internal/models"
	"awesome-alert/internal/types"
	"awesome-alert/pkg/tools"
)

func dbToModel(db *ent.DataSource) *models.DataSource {
	resp := &models.DataSource{
		Name:        db.Name,
		Type:        db.Type,
		Description: db.Description,
		CreatedAt:   db.CreatedAt,
		UpdatedAt:   db.UpdatedAt,
	}

	switch db.Type {
	case types.Prometheus:
		var cfg models.PrometheusConfig
		if err := tools.MapToStruct(db.Config, &cfg); err == nil {
			resp.PrometheusConfig = cfg
		}
	}

	return resp
}

func getConfigMap(req *models.DataSource) (map[string]interface{}, error) {
	switch req.Type {
	case types.Prometheus:
		return tools.StructToMap(req.PrometheusConfig)
	default:
		return nil, nil
	}
}
