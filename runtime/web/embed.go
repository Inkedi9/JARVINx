package web

import "embed"

//go:embed static
var staticFiles embed.FS

func StaticFiles() embed.FS {
	return staticFiles
}
