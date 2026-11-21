// Package web embeds the application views used to discover bootstrap menus.
package web

import "embed"

//go:embed admin/src/views blog/src/views product/src/views
var Views embed.FS
