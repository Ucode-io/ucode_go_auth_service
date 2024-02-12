package cron

import (
	"context"
	"ucode/ucode_go_auth_service/config"
	"ucode/ucode_go_auth_service/pkg/logger"
	"ucode/ucode_go_auth_service/storage"

	"github.com/robfig/cron/v3"
)

type TaskScheduler struct {
	cron *cron.Cron
	cfg  config.Config
	log  logger.LoggerI
	strg storage.StorageI
}

type TaskSchedulerI interface {
	RunJobs(context.Context) error
}

func New(cfg config.Config, log logger.LoggerI, storage storage.StorageI) TaskSchedulerI {
	cron := cron.New()
	defer cron.Start()
	return &TaskScheduler{
		cron: cron,
		cfg:  cfg,
		log:  log,
		strg: storage,
	}
}

func (t *TaskScheduler) RunJobs(ctx context.Context) error {
	t.log.Info("Jobs Started:")
	t.cron.AddFunc("0 0 1 * *", func() {
		t.ApiKeyLimit(context.Background())
	})

	return nil
}

func (t *TaskScheduler) ApiKeyLimit(ctx context.Context) {
	t.log.Info("Started api key monthly request limit job.....")
	err := t.strg.ApiKeys().UpdateIsMonthlyLimitReached(context.Background())
	if err != nil {
		t.log.Error("Error in updating monthly limit reached", logger.Error(err))
		return
	}

	t.log.Info("Finsished api key monthly request limit job.....")
}
