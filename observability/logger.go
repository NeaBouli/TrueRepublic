// Package observability provides the node's defensive logging boundary.
package observability

import sdklog "cosmossdk.io/log"

// Redacted is the only marker used when sensitive log data is removed.
const Redacted = "[REDACTED]"

// Wrap returns a logger that sanitizes messages and structured fields before
// forwarding them. It deliberately does not expose the underlying logger from
// Impl, because doing so would create an unreviewed redaction bypass.
func Wrap(base sdklog.Logger) sdklog.Logger {
	if base == nil {
		base = sdklog.NewNopLogger()
	}
	return redactingLogger{base: base}
}

type redactingLogger struct {
	base sdklog.Logger
}

func (logger redactingLogger) Info(message string, keyValues ...any) {
	logger.base.Info(Sanitize(message), sanitizeKeyValues(keyValues)...)
}

func (logger redactingLogger) Warn(message string, keyValues ...any) {
	logger.base.Warn(Sanitize(message), sanitizeKeyValues(keyValues)...)
}

func (logger redactingLogger) Error(message string, keyValues ...any) {
	logger.base.Error(Sanitize(message), sanitizeKeyValues(keyValues)...)
}

func (logger redactingLogger) Debug(message string, keyValues ...any) {
	logger.base.Debug(Sanitize(message), sanitizeKeyValues(keyValues)...)
}

func (logger redactingLogger) With(keyValues ...any) sdklog.Logger {
	return redactingLogger{base: logger.base.With(sanitizeKeyValues(keyValues)...)}
}

func (logger redactingLogger) Impl() any {
	return logger
}
