package plugin

import "embed"

// Files holds every plugin file a harness needs to load the plugin, so the
// binary can write them out for install. Evals and Go files stay out: the
// harness never reads them. The all: prefix keeps dot dirs like
// .claude-plugin.
//
//go:embed all:.claude-plugin all:hooks all:omp all:output-styles all:references all:skills NOTICE README.md package.json
var Files embed.FS
