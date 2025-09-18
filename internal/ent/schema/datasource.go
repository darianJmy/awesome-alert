package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"time"
)

// DataSource holds the schema definition for the DataSource entity.
type DataSource struct {
	ent.Schema
}

// Fields of the DataSource.
func (DataSource) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			Immutable().
			Comment("数据源名称"),
		field.String("type").
			Immutable().
			Comment("数据源类型"),
		field.String("description").Comment("备注"),
		field.JSON("config", map[string]any{}).
			Optional().
			Comment("数据源配置(JSON)"),
		field.Time("created_at").
			Default(time.Now).
			Immutable().
			Comment("创建日期"),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now).
			Comment("更新时间"),
	}
}

// Edges of the DataSource.
func (DataSource) Edges() []ent.Edge {
	return nil
}
