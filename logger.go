package cobrax

import "log/slog"

var logger *slog.Logger

func init() {
	logger = slog.Default()
}

func SetLogger(l *slog.Logger) {
	logger = l
}
