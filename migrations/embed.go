// Package migrations embeds the SQL migration files into the compiled
// binary so the server can run its own schema migrations on startup,
// without needing the source tree available in the deployment image.
package migrations

import "embed"

//go:embed *.sql
var FS embed.FS
