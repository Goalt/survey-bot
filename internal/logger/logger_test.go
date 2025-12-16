package logger

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestInfo(t *testing.T) {
	ctx := context.Background()

	l := New("test", "info", "unknown", os.Stdout)
	l.Infof(ctx, "test %s", "info")

	ctx = context.WithValue(ctx, UserIDField, "123")
	l.Infof(ctx, "test2 %s", "info")

	ctx = context.WithValue(ctx, UserName, "some_name_123")
	l.Infof(ctx, "test3 %s", "info")
}

func TestWithPrefix(t *testing.T) {
	ctx := context.Background()

	l := New("test", "info", "unknown", os.Stdout).WithPrefix("app-name", "test-app")
	l.Warnf(ctx, "test %s", "info")
}

func TestWithError(t *testing.T) {
	l := New("test", "info", "unknown", os.Stdout).WithPrefix("app-name", "test-app")
	l.WithError(errors.New("some error")).Infof(context.Background(), "test")
}

func TestSentry(t *testing.T) {
	f, err := InitSentry("", 2*time.Second)
	require.NoError(t, err)
	defer f()

	L := New("test", "info", "unknown", os.Stdout)
	var l Logger = L
	l = l.WithPrefix("app-name", "test-app")

	ctx := context.Background()
	ctx = context.WithValue(ctx, UserIDField, "123")
	ctx = context.WithValue(ctx, ChatIDField, "456")
	ctx = context.WithValue(ctx, ReqID, "789an")

	// l.Infof(ctx, "test info %s", "abc")
	// l.Warnf(ctx, "test warn %s", "abc")
	l.WithError(errors.New("some error")).Errorf(ctx, "test error %s", "abc")
}
