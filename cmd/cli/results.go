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
)

type GetResultsCmd struct {
}

func (*GetResultsCmd) Name() string     { return "survey-get-results" }
func (*GetResultsCmd) Synopsis() string { return "get results and print it to output in CSV format" }
func (*GetResultsCmd) Usage() string {
	return `survey-get-results:
	Get results and print it to output in CSV format
  `
}

func (p *GetResultsCmd) SetFlags(f *flag.FlagSet) {
}

func (p *GetResultsCmd) Execute(ctx context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
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

	//todo: use transaction here
	total, err := svc.SaveFinishedSurveys(ctx, nil, os.Stdout, service.ResultsFilter{}, 128)
	if err != nil {
		logger.Errorf(ctx, "failed to save finished surveys: %s", err)
		return subcommands.ExitFailure
	}

	logger.Infof(ctx, "dumped %d surveys states", total)

	return subcommands.ExitSuccess
}
