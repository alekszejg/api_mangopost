package jobs

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/robfig/cron/v3"
)

type Manager struct {
	context context.Context
	jobs    []Job
	cron    *cron.Cron
}

func NewJobManager(ctx context.Context) *Manager {
	return &Manager{
		context: ctx,
		jobs:    []Job{},
		cron:    cron.New(cron.WithSeconds(), cron.WithLocation(time.UTC)),
	}
}

func (jm *Manager) AppendJob(job Job) {
	jm.jobs = append(jm.jobs, job)
}

func (jm *Manager) Stop() {
	shutdownCtx := jm.cron.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	select {
	case <-shutdownCtx.Done():
		slog.Info("All scheduled jobs stopped successfully")
	case <-ctx.Done():
		slog.Warn("Some scheduled jobs may not have completed")
	}
}

func (jm *Manager) ScheduleCronjobs() {
	for _, job := range jm.jobs {
		_, err := jm.cron.AddFunc(job.Schedule(), func() {
			defer func() {
				if r := recover(); r != nil {
					slog.Error(fmt.Sprintf("Cronjob %q panicked: %s", job.Name(), r))
				}
			}()

			ctx, cancel := context.WithTimeout(jm.context, 1*time.Minute)
			defer cancel()

			if err := job.Run(ctx); err != nil {
				slog.Warn(fmt.Sprintf("Cronjob %q failed: %s", job.Name(), err.Error()))
			}
		})

		if err != nil {
			slog.Error(fmt.Sprintf("Failed to schedule %q cronjob: %s", job.Name(), err.Error()))
		}

		slog.Debug(fmt.Sprintf("Successfully scheduled %q cronjob with schedule %q", job.Name(), job.Schedule()))
	}

	jm.cron.Start()
}
