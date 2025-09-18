package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"time"
)

// AlertRule holds the schema definition for the AlertRule entity.
type AlertRule struct {
	ent.Schema
}

// Fields of the AlertRule.
func (AlertRule) Fields() []ent.Field {
	return []ent.Field{
		field.String("group").Comment("分组名称"),
		field.String("name").Comment("规则名称"),
		field.JSON("add_label", map[string]any{}).
			Optional().
			Comment("附加标签"),
		field.String("type").Comment("数据源类型"),
		field.Ints("datasource_id").Comment("数据源ID"),
		field.String("expr").Comment("表达式"),
		field.JSON("condition", map[string]any{}).Comment("触发条件"),
		field.Int("duration").Comment("持续时间"),
		field.Int("alert_continuous_count").Comment("告警连续检测次数"),
		field.Int("recover_continuous_count").Comment("恢复连续检测次数"),
		field.Bool("send_recover_event").Comment("是否推送恢复事件"),
		field.Ints("fault_center").Comment("故障中心"),
		field.Bool("enabled").Comment("是否启用规则"),
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

// Edges of the AlertRule.
func (AlertRule) Edges() []ent.Edge {
	return nil
}
