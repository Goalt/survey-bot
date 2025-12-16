package listener

import (
	"fmt"
	"strconv"
	"time"

	tele "gopkg.in/telebot.v3"

	"git.ykonkov.com/ykonkov/survey-bot/internal/context"
	"git.ykonkov.com/ykonkov/survey-bot/internal/responses"
	"git.ykonkov.com/ykonkov/survey-bot/internal/service"
)

func (l *listener) handleResultsCommand(ctx context.Context, c tele.Context) (err error) {
	l.logger.Infof(ctx, "handle /results command")

	defer func() {
		if errP := recover(); errP != nil {
			err = fmt.Errorf("panic: %v", errP)
		}
	}()

	var f service.ResultsFilter

	switch len(c.Args()) {
	case 2:
		t1, err := time.Parse("2006-01-02", c.Args()[0])
		if err != nil {
			if err := c.Send(responses.InvalidDateFormat); err != nil {
				l.logger.Errorf(ctx, "failed to send message to user: %w", err)
			} else {
				l.logger.Infof(ctx, "send message %s", responses.InvalidDateFormat)
			}
			return nil
		}
		f.From = &t1

		t2, err := time.Parse("2006-01-02", c.Args()[1])
		if err != nil {
			if err := c.Send(responses.InvalidDateFormat); err != nil {
				l.logger.Errorf(ctx, "failed to send message to user: %w", err)
			} else {
				l.logger.Infof(ctx, "send message %s", responses.InvalidDateFormat)
			}
			return nil
		}
		f.To = &t2
	case 1:
		t1, err := time.Parse("2006-01-02", c.Args()[0])
		if err != nil {
			if err := c.Send(responses.InvalidDateFormat); err != nil {
				l.logger.Errorf(ctx, "failed to send message to user: %w", err)
			} else {
				l.logger.Infof(ctx, "send message %s", responses.InvalidDateFormat)
			}
			return nil
		}

		f.From = &t1
	case 0:
		// do nothing
	default:
		if err := c.Send(responses.InvalidNumberOfArguments); err != nil {
			l.logger.Errorf(ctx, "failed to send message to user: %w", err)
		} else {
			l.logger.Infof(ctx, "send message %s", responses.InvalidNumberOfArguments)
		}
		return nil
	}

	return l.svc.HandleResultsCommand(ctx, f)
}

func (l *listener) handleStartCommand(ctx context.Context) (err error) {
	l.logger.Infof(ctx, "handle /start command")

	defer func() {
		if errP := recover(); errP != nil {
			err = fmt.Errorf("panic: %v", errP)
		}
	}()

	return l.svc.HandleStartCommand(ctx)
}

func (l *listener) handleSurveyCommand(ctx context.Context, c tele.Context) (err error) {
	l.logger.Infof(ctx, "handle /survey command")

	defer func() {
		if errP := recover(); errP != nil {
			err = fmt.Errorf("panic: %v", errP)
		}
	}()

	if len(c.Args()) != 1 {
		if err := c.Send(responses.InvalidNumberOfArguments); err != nil {
			l.logger.Errorf(ctx, "failed to send message to user: %w", err)
		} else {
			l.logger.Infof(ctx, "send message ~invalid count of arguments to command~")
		}

		return fmt.Errorf("invalid count of args: %v", c.Args())
	}

	surveyID, err := strconv.ParseInt(c.Args()[0], 10, 64)
	if err != nil {
		if err := c.Send(responses.InvalidSurveyID); err != nil {
			l.logger.Errorf(ctx, "failed to send message to user: %w", err)
		} else {
			l.logger.Infof(ctx, "send message ~invalid survey id~")
		}

		return fmt.Errorf("invalid survey id: %w", err)
	}

	return l.svc.HandleSurveyCommand(ctx, surveyID)
}

func (l *listener) handleListCommand(ctx context.Context) (err error) {
	l.logger.Infof(ctx, "handle /list command")

	defer func() {
		if errP := recover(); errP != nil {
			err = fmt.Errorf("panic: %v", errP)
		}
	}()

	return l.svc.HandleListCommand(ctx)
}

func (l *listener) handleOtherCommand(ctx context.Context, c tele.Context) (err error) {
	l.logger.Infof(ctx, "handle text ")

	defer func() {
		if errP := recover(); errP != nil {
			err = fmt.Errorf("panic: %v", errP)
		}
	}()

	return l.svc.HandleAnswer(ctx, c.Text())
}

func (l *listener) handleMenuCallback(ctx context.Context, c tele.Context) (err error) {
	l.logger.Infof(ctx, "handle menu callback")

	defer func() {
		if err := c.Respond(); err != nil {
			l.logger.Errorf(ctx, "failed to respond to callback: %w", err)
		}
	}()

	defer func() {
		if errP := recover(); errP != nil {
			err = fmt.Errorf("panic: %v", errP)
		}
	}()

	callback := c.Callback()
	if callback == nil {
		return fmt.Errorf("callback is nil")
	}

	switch callback.Unique {
	case "survey_id_1":
		err = l.svc.HandleSurveyCommand(ctx, 1)
	case "survey_id_2":
		err = l.svc.HandleSurveyCommand(ctx, 2)
	case "survey_id_3":
		err = l.svc.HandleSurveyCommand(ctx, 3)
	case "survey_id_4":
		err = l.svc.HandleSurveyCommand(ctx, 4)
	case "survey_id_5":
		err = l.svc.HandleSurveyCommand(ctx, 5)
	case "survey_id_6":
		err = l.svc.HandleSurveyCommand(ctx, 6)
	default:
		return fmt.Errorf("unknown callback: %v", c.Callback().Unique)
	}

	if err != nil {
		return fmt.Errorf("failed to handle callback: %w", err)
	}

	return c.Respond()
}

func (l *listener) handleAnswerCallback(ctx context.Context, c tele.Context) (err error) {
	l.logger.Infof(ctx, "handle answer callback")

	defer func() {
		if err := c.Respond(); err != nil {
			l.logger.Errorf(ctx, "failed to respond to callback: %w", err)
		}
	}()

	defer func() {
		if errP := recover(); errP != nil {
			err = fmt.Errorf("panic: %v", errP)
		}
	}()

	callback := c.Callback()
	if callback == nil {
		return fmt.Errorf("callback is nil")
	}

	switch callback.Unique {
	case "answer_1":
		err = l.svc.HandleAnswer(ctx, "1")
	case "answer_2":
		err = l.svc.HandleAnswer(ctx, "2")
	case "answer_3":
		err = l.svc.HandleAnswer(ctx, "3")
	case "answer_4":
		err = l.svc.HandleAnswer(ctx, "4")
	case "answer_5":
		err = l.svc.HandleAnswer(ctx, "5")
	case "answer_6":
		err = l.svc.HandleAnswer(ctx, "6")
	case "answer_7":
		err = l.svc.HandleAnswer(ctx, "7")
	case "answer_8":
		err = l.svc.HandleAnswer(ctx, "8")
	case "answer_9":
		err = l.svc.HandleAnswer(ctx, "9")
	case "answer_10":
		err = l.svc.HandleAnswer(ctx, "10")
	case "answer_11":
		err = l.svc.HandleAnswer(ctx, "11")
	default:
		return fmt.Errorf("unknown callback: %v", c.Callback().Unique)
	}

	if err != nil {
		return fmt.Errorf("failed to handle callback: %w", err)
	}

	return c.Respond()
}

func (l *listener) handleListOfSurveyCallback(ctx context.Context, c tele.Context) (err error) {
	l.logger.Infof(ctx, "handle list of survey callback")

	defer func() {
		if err := c.Respond(); err != nil {
			l.logger.Errorf(ctx, "failed to respond to callback: %w", err)
		}
	}()

	defer func() {
		if errP := recover(); errP != nil {
			err = fmt.Errorf("panic: %v", errP)
		}
	}()

	callback := c.Callback()
	if callback == nil {
		return fmt.Errorf("callback is nil")
	}

	switch callback.Unique {
	case "menu":
		err = l.svc.HandleListCommand(ctx)
	default:
		return fmt.Errorf("unknown callback: %v", c.Callback().Unique)
	}

	if err != nil {
		return fmt.Errorf("failed to handle callback: %w", err)
	}

	return c.Respond()
}
