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

type CreateSurveyCmd struct {
}

func (*CreateSurveyCmd) Name() string     { return "survey-create" }
func (*CreateSurveyCmd) Synopsis() string { return "Create survey from file" }
func (*CreateSurveyCmd) Usage() string {
	return `survey-create <file_path>:
	Creates survey and prints guid and id of survey
  `
}

func (p *CreateSurveyCmd) SetFlags(f *flag.FlagSet) {
}

func (p *CreateSurveyCmd) Execute(ctx context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	filename := f.Arg(0)

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

	survey, err := service.ReadSurveyFromFile(filename)
	if err != nil {
		logger.Errorf(ctx, "failed to read survey from file: %s", err)
		return subcommands.ExitFailure
	}

	repo := db.New(sqlDB)
	processor := resultsprocessor.New()
	svc := service.New(nil, repo, processor, logger)

	survey, err = svc.CreateSurvey(context.Background(), survey)
	if err != nil {
		logger.Errorf(ctx, "failed to create survey: %s", err)
		return subcommands.ExitFailure
	}

	logger.Infof(ctx, "survey created with id %d and guid %s", survey.ID, survey.GUID)

	return subcommands.ExitSuccess
}

type UpdateSurveyCmd struct {
}

func (*UpdateSurveyCmd) Name() string               { return "survey-update" }
func (p *UpdateSurveyCmd) SetFlags(f *flag.FlagSet) {}
func (*UpdateSurveyCmd) Synopsis() string {
	return "Update survey from file"
}
func (*UpdateSurveyCmd) Usage() string {
	return `survey-update <survey_guid> <file_path>:
	Updates survey from file by guid. Updates "name", "description", "questions" and "calculations_type" fields.
  `
}

func (p *UpdateSurveyCmd) Execute(ctx context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	config, err := config.New()
	if err != nil {
		log.Print("failed to read config: ", err)
		return subcommands.ExitFailure
	}

	logger := logger.New(config.Level, config.Env, config.ReleaseVersion, os.Stdout)

	surveyGUIDRaw := f.Arg(0)
	filename := f.Arg(1)

	surveyGUID, err := uuid.Parse(surveyGUIDRaw)
	if err != nil {
		log.Print("failed to parse survey guid: ", err)
		return subcommands.ExitFailure
	}

	sqlDB, err := db.ConnectWithTimeout(time.Minute, config.DB)
	if err != nil {
		logger.Errorf(ctx, "failed to connect to db: %s", err)
		return subcommands.ExitFailure
	}

	new, err := service.ReadSurveyFromFile(filename)
	if err != nil {
		logger.Errorf(ctx, "failed to read survey from file: %s", err)
		return subcommands.ExitFailure
	}

	repo := db.New(sqlDB)
	processor := resultsprocessor.New()
	svc := service.New(nil, repo, processor, logger)

	new.GUID = surveyGUID

	if err := svc.UpdateSurvey(ctx, new); err != nil {
		logger.Errorf(ctx, "failed to update survey: %s", err)
		return subcommands.ExitFailure
	}

	logger.Infof(ctx, "survey updated with guid %s", new.GUID)

	return subcommands.ExitSuccess
}
