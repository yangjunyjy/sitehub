// Package gatewayadmin serves the compiled Vue app embedded in the gateway.
package gatewayadmin

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:generate npm run build
//go:embed static
var assets embed.FS

func Handler() http.Handler {
	root, err := fs.Sub(assets, "static")
	if err != nil {
		panic(err)
	}
	return http.FileServer(http.FS(root))
}
