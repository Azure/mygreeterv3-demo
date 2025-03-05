package logattrs

import (
	log "log/slog"
)

var attrs []log.Attr

func init() {
}

func GetAttrs() []log.Attr {
	return attrs
}
