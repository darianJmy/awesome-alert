package ctx

import (
	"awesome-alert/internal/ent"
	"context"
)

type Context struct {
	DB  *ent.Client
	Ctx context.Context
}

func NewContext(ctx context.Context, db *ent.Client) *Context {
	return &Context{
		DB:  db,
		Ctx: ctx,
	}
}
