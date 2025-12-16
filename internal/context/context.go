package context

import (
	stdcontext "context"
	"fmt"

	"git.ykonkov.com/ykonkov/survey-bot/internal/logger"
	tele "gopkg.in/telebot.v3"
)

type (
	Botter interface {
		Send(interface{}, ...interface{}) error
		Sender() *tele.User
		Chat() *tele.Chat
	}

	Context interface {
		Send(msg interface{}, options ...interface{}) error
		UserID() int64
		ChatID() int64
		Nickname() string
		SetStdContext(stdcontext.Context)
		stdcontext.Context
	}

	context struct {
		stdcontext.Context
		b Botter
	}
)

func New(ctx stdcontext.Context, b Botter, id string) *context {
	ctx = stdcontext.WithValue(ctx, logger.UserIDField, b.Sender().ID)
	ctx = stdcontext.WithValue(ctx, logger.ChatIDField, b.Chat().ID)
	ctx = stdcontext.WithValue(ctx, logger.ReqID, id)
	ctx = stdcontext.WithValue(
		ctx,
		logger.UserName,
		fmt.Sprintf("%s %s (%s)", b.Sender().FirstName, b.Sender().LastName, b.Sender().Username),
	)

	return &context{
		b:       b,
		Context: ctx,
	}
}

func (c *context) Send(msg interface{}, options ...interface{}) error {
	return c.b.Send(msg, options...)
}

func (c *context) UserID() int64 {
	return c.b.Sender().ID
}

func (c *context) ChatID() int64 {
	return c.b.Chat().ID
}

func (c *context) Nickname() string {
	return fmt.Sprintf("%s %s (%s)", c.b.Sender().FirstName, c.b.Sender().LastName, c.b.Sender().Username)
}

func (c *context) SetStdContext(ctx stdcontext.Context) {
	c.Context = ctx
}
