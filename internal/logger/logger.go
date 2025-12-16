package logger

import (
	"context"
	"fmt"
	"io"
	"runtime"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/sirupsen/logrus"
)

const (
	UserIDField Field = "user_id"
	ChatIDField Field = "chat_id"
	ReqID       Field = "req_id"
	UserName    Field = "user_name"

	sentryID = "sentry_id"
)

var ctxFields = map[Field]any{
	UserIDField: nil,
	ChatIDField: nil,
	ReqID:       nil,
	UserName:    nil,
}

type (
	Logger interface {
		Infof(ctx context.Context, format string, args ...any)
		Warnf(ctx context.Context, format string, args ...any)
		Errorf(ctx context.Context, format string, args ...any)
		WithPrefix(k, v string) Logger
		WithError(err error) Logger
	}

	logger struct {
		l              *logrus.Entry
		env            string
		releaseVersion string
	}

	Field string
)

var _ Logger = &logger{}

func New(env string, level string, releaseVersion string, output io.Writer) *logger {
	l := logrus.New()
	l.SetOutput(output)
	l.SetFormatter(&logrus.JSONFormatter{})

	switch level {
	case "info":
		l.SetLevel(logrus.InfoLevel)
	case "warn":
		l.SetLevel(logrus.WarnLevel)
	case "error":
		l.SetLevel(logrus.ErrorLevel)
	default:
		l.SetLevel(logrus.InfoLevel)
	}

	lEntry := logrus.NewEntry(l)
	lEntry = lEntry.WithField("env", env)
	lEntry = lEntry.WithField("release_version", releaseVersion)
	return &logger{lEntry, env, releaseVersion}
}

func (l *logger) Infof(ctx context.Context, format string, args ...any) {
	fields := getFieldsFromContext(ctx)
	l2 := l.l.WithFields(fields)
	fields = l2.Data

	id := l.captureEvent(ctx, sentry.LevelInfo, fields, format, args...)
	l2.WithField(sentryID, id).Infof(format, args...)
}

func (l *logger) Warnf(ctx context.Context, format string, args ...any) {
	fields := getFieldsFromContext(ctx)
	l2 := l.l.WithFields(fields)
	fields = l2.Data

	id := l.captureEvent(ctx, sentry.LevelWarning, fields, format, args...)
	l2.WithField(sentryID, id).Warnf(format, args...)
}

func (l *logger) Errorf(ctx context.Context, format string, args ...any) {
	fields := getFieldsFromContext(ctx)
	l2 := l.l.WithFields(fields)
	fields = l2.Data

	id := l.captureEvent(ctx, sentry.LevelError, fields, format, args...)
	l2.WithField(sentryID, id).Errorf(format, args...)
}

func (l *logger) WithPrefix(k, v string) Logger {
	return &logger{l.l.WithField(k, v), l.env, l.releaseVersion}
}

func (l *logger) WithError(err error) Logger {
	return &logger{l.l.WithError(err), l.env, l.releaseVersion}
}

func getFieldsFromContext(ctx context.Context) logrus.Fields {
	result := make(logrus.Fields)
	for fieldName := range ctxFields {
		if val := ctx.Value(fieldName); val != nil {
			result[string(fieldName)] = val
		}
	}

	// set report caller
	_, file, line, ok := runtime.Caller(2)
	if !ok {
		result["func"] = "unknown"
		return result
	}

	val := fmt.Sprintf("%s:%d", file, line)
	result["func"] = val

	return result
}

func InitSentry(dsn string, flushTime time.Duration) (func(), error) {
	if dsn == "" {
		return func() {}, nil
	}

	err := sentry.Init(sentry.ClientOptions{
		Dsn:              dsn,
		EnableTracing:    true,
		TracesSampleRate: 1.0,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to init sentry: %w", err)
	}

	return func() {
		sentry.Flush(flushTime)
	}, nil
}

func (l *logger) captureEvent(
	ctx context.Context,
	level sentry.Level,
	fields logrus.Fields,
	format string,
	args ...any,
) string {
	if sentry.CurrentHub() == nil {
		return ""
	}

	event := sentry.NewEvent()
	event.Level = level
	event.Message = fmt.Sprintf(format, args...)

	fieldsFormatted := map[string]interface{}{}
	for k, v := range fields {
		if err, ok := v.(error); ok {
			fieldsFormatted[k] = err.Error()
			continue
		}

		fieldsFormatted[k] = v
	}

	transaction := sentry.TransactionFromContext(ctx)
	if transaction != nil {
		transaction.Data = fields
		transaction.Description = fmt.Sprintf(format, args...)
		event.Transaction = transaction.TraceID.String()
	}

	event.Extra = fieldsFormatted
	event.Release = l.releaseVersion
	event.Environment = l.env

	if val, ok := fields[string(UserIDField)].(string); ok {
		event.User.ID = val
	}

	id := sentry.CaptureEvent(event)
	if id == nil {
		return ""
	}

	return string(*id)
}
