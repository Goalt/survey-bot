package main

import (
	"context"
	"flag"
	"log"
	"os"
	"time"

	"git.ykonkov.com/ykonkov/survey-bot/internal/config"
	"git.ykonkov.com/ykonkov/survey-bot/internal/logger"
	"git.ykonkov.com/ykonkov/survey-bot/internal/repository/db"
	"git.ykonkov.com/ykonkov/survey-bot/internal/resultsprocessor"
	"git.ykonkov.com/ykonkov/survey-bot/internal/service"
	"github.com/google/subcommands"
	"github.com/google/uuid"
)

type DeleteUserInfoCmd struct {
}

func (*DeleteUserInfoCmd) Name() string     { return "survey-state-delete" }
func (*DeleteUserInfoCmd) Synopsis() string { return "delete survey state from file" }
func (*DeleteUserInfoCmd) Usage() string {
	return `survey-state-delete <user_guid> <survey_guid>:
	Delete survey state
  `
}

func (p *DeleteUserInfoCmd) SetFlags(f *flag.FlagSet) {
}

func (p *DeleteUserInfoCmd) Execute(ctx context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	var (
		userGUID, surveyGUID uuid.UUID
		err                  error
	)
	if userGUID, err = uuid.Parse(f.Arg(0)); err != nil {
		log.Print("failed to parse user guid: ", err)
		return subcommands.ExitFailure
	}

	if surveyGUID, err = uuid.Parse(f.Arg(1)); err != nil {
		log.Print("failed to parse survey guid: ", err)
		return subcommands.ExitFailure
	}

	config, err := config.New()
	if err != nil {
		log.Print("failed to read config: ", err)
		return subcommands.ExitFailure
	}

	logger := logger.New(config.Level, config.Env, config.ReleaseVersion, os.Stdout)

	sqlDB, err := db.ConnectWithTimeout(time.Minute, config.DB)
	if err != nil {
		logger.Errorf(ctx, "failed to connect to db: %s", err)
		return subcommands.ExitFailure
	}

	repo := db.New(sqlDB)
	processor := resultsprocessor.New()
	svc := service.New(nil, repo, processor, logger)

	// get user by user_guid
	user, err := svc.GetUserByGUID(ctx, userGUID)
	if err != nil {
		logger.Errorf(ctx, "failed to get user by guid: %s", err)
		return subcommands.ExitFailure
	}

	// if user's current survey is equal to provided in arguments, set it to null
	if user.CurrentSurvey != nil {
		if err := svc.SetUserCurrentSurveyToNil(ctx, userGUID); err != nil {
			logger.Errorf(ctx, "failed to update user: %s", err)
			return subcommands.ExitFailure
		}
	}

	// delete where user_guid and survey_guid are equal to arguments
	if err := svc.DeleteUserSurvey(ctx, userGUID, surveyGUID); err != nil {
		logger.Errorf(ctx, "failed to delete user survey: %s", err)
		return subcommands.ExitFailure
	}

	logger.Infof(ctx, "user survey deleted with user_guid %s and survey_guid %s", userGUID, surveyGUID)

	return subcommands.ExitSuccess
}
