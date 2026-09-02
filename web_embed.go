//go:build !lite

package main

import (
	"embed"
)

//go:embed all:clients/player-build/web-embedded
var WebDist embed.FS
