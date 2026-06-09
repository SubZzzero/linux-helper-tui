package linuxhelper

import "embed"

// Assets holds all bundled recipes, locales, and themes.
//
//go:embed assets
var Assets embed.FS
