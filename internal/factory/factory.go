package factory

import (
	"awesome-alert/internal/ctx"
	"awesome-alert/internal/factory/alertgroup"
	"awesome-alert/internal/factory/alertrule"
	"awesome-alert/internal/factory/datasource"
)

var Factory *factory

type Interface interface {
	DataSource() datasource.Interface
	AlertGroup() alertgroup.Interface
}

type factory struct {
	datasource *datasource.DataSource
	alertGroup *alertgroup.AlertGroup
	alertRule  *alertrule.AlertRule
}

func (f *factory) DataSource() datasource.Interface {
	return f.datasource
}

func (f *factory) AlertGroup() alertgroup.Interface {
	return f.alertGroup
}

func (f *factory) AlertRule() alertrule.Interface {
	return f.alertRule
}

func NewFactory(ctx *ctx.Context) {
	var fc factory

	fc.datasource = datasource.NewDataSource(ctx)
	fc.alertGroup = alertgroup.NewAlertGroup(ctx)
	fc.alertRule = alertrule.NewAlertRule(ctx)

	Factory = &fc
}
