// Package webassets embeds the built frontend in the server binary.
package webassets

import "embed"

//go:embed dist/*
var Files embed.FS
