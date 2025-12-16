package listener

import (
	"context"
	"slices"

	"git.ykonkov.com/ykonkov/survey-bot/internal/logger"
	tele "gopkg.in/telebot.v3"
)

func NewAdminMiddleware(adminUserID []int64, log logger.Logger) tele.MiddlewareFunc {
	return func(next tele.HandlerFunc) tele.HandlerFunc {
		return func(c tele.Context) error {
			if !slices.Contains(adminUserID, c.Sender().ID) {
				log.Warnf(context.Background(), "user %v is not admin", c.Sender().ID)
				return nil
			}

			return next(c) // continue execution chain
		}
	}
}
