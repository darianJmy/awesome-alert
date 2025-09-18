package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"time"
)

// AlertGroup holds the schema definition for the AlertGroup entity.
type AlertGroup struct {
	ent.Schema
}

// Fields of the AlertGroup.
func (AlertGroup) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			Immutable().
			Comment("分组名称"),
		field.String("description").Comment("备注"),
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

// Edges of the AlertGroup.
func (AlertGroup) Edges() []ent.Edge {
	return nil
}
