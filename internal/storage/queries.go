package storage

import "embed"

//go:embed queries/*.sql
var queriesFS embed.FS
