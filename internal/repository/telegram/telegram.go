package telegram

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/getsentry/sentry-go"
	"github.com/prometheus/client_golang/prometheus"
	tele "gopkg.in/telebot.v3"

	"git.ykonkov.com/ykonkov/survey-bot/internal/context"
	"git.ykonkov.com/ykonkov/survey-bot/internal/entity"
	"git.ykonkov.com/ykonkov/survey-bot/internal/responses"
	"git.ykonkov.com/ykonkov/survey-bot/internal/service"
)

type client struct {
}

func NewClient() *client {
	return &client{}
}

func (c *client) SendSurveyList(ctx context.Context, states []service.UserSurveyState) error {
	span := sentry.StartSpan(ctx, "SendSurveyList")
	defer span.Finish()

	// Inline buttons.
	//
	// Pressing it will cause the client to
	// send the bot a callback.
	//
	// Make sure Unique stays unique as per button kind
	// since it's required for callback routing to work.
	//
	selector := &tele.ReplyMarkup{}
	var rows []tele.Row

	builder := strings.Builder{}
	builder.WriteString(fmt.Sprintf("%s:\n", responses.ChooseSurvey))
	for _, state := range states {
		switch {
		case state.IsCurrent:
			builder.WriteString(fmt.Sprintf("%d - %s (текущий)", state.Survey.ID, state.Survey.Name))
			rows = append(rows, selector.Row(selector.Data(state.Survey.Name, "survey_id_"+strconv.FormatInt(state.Survey.ID, 10))))
		case state.State == entity.FinishedState:
			builder.WriteString(fmt.Sprintf("%d - %s (завершен)", state.Survey.ID, state.Survey.Name))
		case state.State == entity.ActiveState:
			builder.WriteString(fmt.Sprintf("%d - %s (в процессе)", state.Survey.ID, state.Survey.Name))
			rows = append(rows, selector.Row(selector.Data(state.Survey.Name, "survey_id_"+strconv.FormatInt(state.Survey.ID, 10))))
		case state.State == entity.NotStartedState:
			builder.WriteString(fmt.Sprintf("%d - %s", state.Survey.ID, state.Survey.Name))
			rows = append(rows, selector.Row(selector.Data(state.Survey.Name, "survey_id_"+strconv.FormatInt(state.Survey.ID, 10))))
		default:
			return fmt.Errorf("unknown state: %v", state.State)
		}

		builder.WriteString("\n")
	}

	selector.Inline(
		rows...,
	)

	timer := prometheus.NewTimer(messageDuration.WithLabelValues("SendSurveyList"))
	defer timer.ObserveDuration()

	if err := ctx.Send(builder.String(), selector); err != nil {
		messageCounter.WithLabelValues("failed", "SendSurveyList").Inc()
		return fmt.Errorf("failed to send msg: %w", err)
	}

	messageCounter.WithLabelValues("success", "SendSurveyList").Inc()

	return nil
}

func (c *client) SendSurveyQuestion(ctx context.Context, question entity.Question) error {
	span := sentry.StartSpan(ctx, "SendSurveyQuestion")
	defer span.Finish()

	builder := strings.Builder{}
	builder.WriteString(fmt.Sprintf("Вопрос: %s\n", question.Text))

	selector := &tele.ReplyMarkup{}
	var rows []tele.Row

	switch question.AnswerType {
	case entity.AnswerTypeSegment:
		builder.WriteString(fmt.Sprintf("Напишите число из диапазона: %d - %d\n", question.PossibleAnswers[0], question.PossibleAnswers[1]))

	case entity.AnswerTypeSelect:
		builder.WriteString("Выберите один из вариантов:\n")

		for i := range question.PossibleAnswers {
			builder.WriteString(fmt.Sprintf("%d - %s", question.PossibleAnswers[i], question.AnswersText[i]))
			builder.WriteString("\n")
			rows = append(
				rows,
				selector.Row(
					selector.Data(
						strconv.FormatInt(int64(question.PossibleAnswers[i]), 10),
						"answer_"+strconv.FormatInt(int64(question.PossibleAnswers[i]), 10),
					),
				),
			)
		}

	case entity.AnswerTypeMultiSelect:
		builder.WriteString("Напишите один или несколько вариантов через запятую:\n")

		for i := range question.PossibleAnswers {
			builder.WriteString(fmt.Sprintf("%d - %s", question.PossibleAnswers[i], question.AnswersText[i]))
			builder.WriteString("\n")
		}
	default:
		return fmt.Errorf("unknown answer type: %v", question.AnswerType)
	}

	rows = append(rows, selector.Row(selector.Data("Назад к списку тестов", "menu")))

	selector.Inline(
		rows...,
	)

	timer := prometheus.NewTimer(messageDuration.WithLabelValues("SendSurveyQuestion"))
	defer timer.ObserveDuration()

	if err := ctx.Send(builder.String(), selector); err != nil {
		messageCounter.WithLabelValues("failed", "SendSurveyQuestion").Inc()
		return fmt.Errorf("failed to send msg: %w", err)
	}

	messageCounter.WithLabelValues("success", "SendSurveyQuestion").Inc()

	return nil
}

func (c *client) SendMessage(ctx context.Context, msg string) error {
	span := sentry.StartSpan(ctx, "SendMessage")
	defer span.Finish()

	selector := &tele.ReplyMarkup{}
	var rows []tele.Row

	rows = append(rows, selector.Row(selector.Data("Назад к списку тестов", "menu")))

	selector.Inline(
		rows...,
	)

	timer := prometheus.NewTimer(messageDuration.WithLabelValues("SendMessage"))
	defer timer.ObserveDuration()

	if err := ctx.Send(msg, selector); err != nil {
		messageCounter.WithLabelValues("failed", "SendMessage").Inc()
		return fmt.Errorf("failed to send msg: %w", err)
	}

	messageCounter.WithLabelValues("success", "SendMessage").Inc()

	return nil
}

func (c *client) SendFile(ctx context.Context, path string) error {
	span := sentry.StartSpan(ctx, "SendFile")
	defer span.Finish()

	a := &tele.Document{
		File:     tele.FromDisk(path),
		FileName: "results.csv",
	}

	timer := prometheus.NewTimer(messageDuration.WithLabelValues("SendFile"))
	defer timer.ObserveDuration()

	// Will upload the file from disk and send it to the recipient
	if err := ctx.Send(a); err != nil {
		messageCounter.WithLabelValues("failed", "SendFile").Inc()
		return fmt.Errorf("failed to send file: %w", err)
	}

	messageCounter.WithLabelValues("success", "SendFile").Inc()

	return nil
}
