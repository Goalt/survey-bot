package listener

import (
	stdcontext "context"
	"fmt"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/prometheus/client_golang/prometheus"
	tele "gopkg.in/telebot.v3"

	"git.ykonkov.com/ykonkov/survey-bot/internal/context"
	"git.ykonkov.com/ykonkov/survey-bot/internal/logger"
	"git.ykonkov.com/ykonkov/survey-bot/internal/service"
)

type (
	Listener interface {
		Start()
		Stop()
	}

	listener struct {
		logger logger.Logger
		b      *tele.Bot
		svc    service.Service
	}
)

func New(
	logger logger.Logger,
	token string,
	adminUserIDs []int64,
	pollInterval time.Duration,
	svc service.Service,
) (*listener, error) {
	l := &listener{
		logger: logger,
		svc:    svc,
	}

	pref := tele.Settings{
		Token:  token,
		Poller: &tele.LongPoller{Timeout: pollInterval},
	}

	b, err := tele.NewBot(pref)
	if err != nil {
		return nil, fmt.Errorf("failed to create tele bot: %w", err)
	}

	b.Handle("/results", func(c tele.Context) error {
		span := l.initSentryContext(stdcontext.Background(), "handleResultsCommand")
		defer span.Finish()
		ctx := context.New(span.Context(), c, span.TraceID.String())

		timer := prometheus.NewTimer(listenerDuration.WithLabelValues("handleResultsCommand"))
		defer timer.ObserveDuration()

		if err := l.handleResultsCommand(ctx, c); err != nil {
			listenerCounter.WithLabelValues("failed", "handleResultsCommand").Inc()
			l.logger.WithError(err).Errorf(ctx, "failed to handle /results command")
		} else {
			listenerCounter.WithLabelValues("success", "handleResultsCommand").Inc()
		}

		return nil
	}, NewAdminMiddleware(adminUserIDs, logger))

	b.Handle("/start", func(c tele.Context) error {
		span := l.initSentryContext(stdcontext.Background(), "handleStartCommand")
		defer span.Finish()
		ctx := context.New(span.Context(), c, span.TraceID.String())

		timer := prometheus.NewTimer(listenerDuration.WithLabelValues("handleStartCommand"))
		defer timer.ObserveDuration()

		if err := l.handleStartCommand(ctx); err != nil {
			listenerCounter.WithLabelValues("failed", "handleStartCommand").Inc()
			l.logger.WithError(err).Errorf(ctx, "failed to handle /start command")
		} else {
			listenerCounter.WithLabelValues("success", "handleStartCommand").Inc()
		}

		return nil
	})

	b.Handle("/survey", func(c tele.Context) error {
		span := l.initSentryContext(stdcontext.Background(), "handleSurveyCommand")
		defer span.Finish()
		ctx := context.New(span.Context(), c, span.TraceID.String())

		timer := prometheus.NewTimer(listenerDuration.WithLabelValues("handleSurveyCommand"))
		defer timer.ObserveDuration()

		if err := l.handleSurveyCommand(ctx, c); err != nil {
			listenerCounter.WithLabelValues("failed", "handleSurveyCommand").Inc()
			l.logger.WithError(err).Errorf(ctx, "failed to handle /survey command")
		} else {
			listenerCounter.WithLabelValues("success", "handleSurveyCommand").Inc()
		}

		return nil
	})

	b.Handle("/list", func(c tele.Context) error {
		span := l.initSentryContext(stdcontext.Background(), "handleListCommand")
		defer span.Finish()
		ctx := context.New(span.Context(), c, span.TraceID.String())

		timer := prometheus.NewTimer(listenerDuration.WithLabelValues("handleListCommand"))
		defer timer.ObserveDuration()

		if err := l.handleListCommand(ctx); err != nil {
			listenerCounter.WithLabelValues("failed", "handleListCommand").Inc()
			l.logger.WithError(err).Errorf(ctx, "failed to handle /list command")
		} else {
			listenerCounter.WithLabelValues("success", "handleListCommand").Inc()
		}

		return nil
	})

	b.Handle(tele.OnText, func(c tele.Context) error {
		span := l.initSentryContext(stdcontext.Background(), "handleOtherCommand")
		defer span.Finish()
		ctx := context.New(span.Context(), c, span.TraceID.String())

		timer := prometheus.NewTimer(listenerDuration.WithLabelValues("handleOtherCommand"))
		defer timer.ObserveDuration()

		if err := l.handleOtherCommand(ctx, c); err != nil {
			listenerCounter.WithLabelValues("failed", "handleOtherCommand").Inc()
			l.logger.WithError(err).Errorf(ctx, "failed to handle message")
		} else {
			listenerCounter.WithLabelValues("success", "handleOtherCommand").Inc()
		}

		return nil
	})

	selector := &tele.ReplyMarkup{}
	var menuButtons = []tele.Btn{
		selector.Data("", "survey_id_1"),
		selector.Data("", "survey_id_2"),
		selector.Data("", "survey_id_3"),
		selector.Data("", "survey_id_4"),
		selector.Data("", "survey_id_5"),
		selector.Data("", "survey_id_6"),
	}
	var answerButtons = []tele.Btn{
		selector.Data("", "answer_1"),
		selector.Data("", "answer_2"),
		selector.Data("", "answer_3"),
		selector.Data("", "answer_4"),
		selector.Data("", "answer_5"),
		selector.Data("", "answer_6"),
		selector.Data("", "answer_7"),
		selector.Data("", "answer_8"),
		selector.Data("", "answer_9"),
		selector.Data("", "answer_10"),
		selector.Data("", "answer_11"),
	}
	listOfSurveysBtn := selector.Data("", "menu")

	b.Handle(&listOfSurveysBtn, func(c tele.Context) error {
		span := l.initSentryContext(stdcontext.Background(), "handleListOfSurveyCallback")
		defer span.Finish()
		ctx := context.New(span.Context(), c, span.TraceID.String())

		timer := prometheus.NewTimer(listenerDuration.WithLabelValues("handleListOfSurveyCallback"))
		defer timer.ObserveDuration()

		if err := l.handleListOfSurveyCallback(ctx, c); err != nil {
			listenerCounter.WithLabelValues("failed", "handleListOfSurveyCallback").Inc()
			l.logger.WithError(err).Errorf(ctx, "failed to handle menu callback")
		} else {
			listenerCounter.WithLabelValues("success", "handleListOfSurveyCallback").Inc()
		}

		return nil
	})

	// On menu button pressed (callback)
	for _, btn := range menuButtons {
		b.Handle(&btn, func(c tele.Context) error {
			span := l.initSentryContext(stdcontext.Background(), "handleMenuCallback")
			defer span.Finish()
			ctx := context.New(span.Context(), c, span.TraceID.String())

			timer := prometheus.NewTimer(listenerDuration.WithLabelValues("handleMenuCallback"))
			defer timer.ObserveDuration()

			if err := l.handleMenuCallback(ctx, c); err != nil {
				listenerCounter.WithLabelValues("failed", "handleMenuCallback").Inc()
				l.logger.WithError(err).Errorf(ctx, "failed to handle menu callback")
			} else {
				listenerCounter.WithLabelValues("success", "handleMenuCallback").Inc()
			}

			return nil
		})
	}

	for _, btn := range answerButtons {
		b.Handle(&btn, func(c tele.Context) error {
			span := l.initSentryContext(stdcontext.Background(), "handleAnswerCallback")
			defer span.Finish()
			ctx := context.New(span.Context(), c, span.TraceID.String())

			timer := prometheus.NewTimer(listenerDuration.WithLabelValues("handleAnswerCallback"))
			defer timer.ObserveDuration()

			if err := l.handleAnswerCallback(ctx, c); err != nil {
				listenerCounter.WithLabelValues("failed", "handleAnswerCallback").Inc()
				l.logger.WithError(err).Errorf(ctx, "failed to handle menu callback")
			} else {
				listenerCounter.WithLabelValues("success", "handleAnswerCallback").Inc()
			}

			return nil
		})
	}

	l.b = b

	return l, nil
}

func (l *listener) Start() {
	l.b.Start()
}

func (l *listener) Stop() {
	l.b.Stop()
}

func (l *listener) initSentryContext(ctx stdcontext.Context, name string) *sentry.Span {
	hub := sentry.GetHubFromContext(ctx)
	if hub == nil {
		// Check the concurrency guide for more details: https://docs.sentry.io/platforms/go/concurrency/
		hub = sentry.CurrentHub().Clone()
		ctx = sentry.SetHubOnContext(ctx, hub)
	}

	options := []sentry.SpanOption{
		// Set the OP based on values from https://develop.sentry.dev/sdk/performance/span-operations/
		sentry.WithOpName("listener"),
		sentry.WithTransactionSource(sentry.SourceURL),
	}

	return sentry.StartTransaction(ctx,
		name,
		options...,
	)
}
