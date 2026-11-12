package web

import (
	"embed"
	"io/fs"
)

//go:embed all:dist
var distFS embed.FS

// Dist returns the filesystem representing the compiled React visual studio.
func Dist() (fs.FS, error) {
	return fs.Sub(distFS, "dist")
}
