package hardloop

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"
)

var ErrCronAlreadyRunning = errors.New("cron already running")

type cronJob struct {
	Jobs []Cron

	started bool
	m       sync.Mutex
	wg      sync.WaitGroup
	cancel  context.CancelFunc
	log     Logger
}

type Cron struct {
	// Name for logging and identification purposes.
	Name  string
	Func  func(ctx context.Context) error
	Specs []string

	schedules []Schedule
}

func NewCron(crons ...Cron) (*cronJob, error) {
	jobs := make([]Cron, 0, len(crons))
	for _, cron := range crons {
		schedules := make([]Schedule, 0, len(cron.Specs))
		for _, spec := range cron.Specs {
			startSchedule, err := ParseStandard(spec)
			if err != nil {
				return nil, err
			}

			schedules = append(schedules, startSchedule)
		}

		if len(schedules) == 0 {
			continue
		}

		jobs = append(jobs, Cron{
			Name:      cron.Name,
			Func:      cron.Func,
			Specs:     cron.Specs,
			schedules: schedules,
		})
	}

	return &cronJob{
		Jobs: jobs,
		log:  slog.Default(),
	}, nil
}

func (c *cronJob) SetLogger(log Logger) {
	c.log = log
}

// Start starts the cron job, running each job according to its schedule.
//   - If the cron job is already running, it returns an error.
func (c *cronJob) Start(ctx context.Context) error {
	if ok := c.m.TryLock(); !ok {
		return ErrCronAlreadyRunning
	}
	defer c.m.Unlock()

	if c.started {
		return ErrCronAlreadyRunning
	}
	c.started = true

	ctx, cancel := context.WithCancel(ctx)
	c.cancel = cancel

	for _, job := range c.Jobs {
		c.wg.Add(1)

		if c.log != nil {
			c.log.Info("add cron job", "job", job.Name, "specs", job.Specs)
		}

		go func(job Cron) {
			defer c.wg.Done()

			var nextTime time.Time
			for {
				now := time.Now()
				if !nextTime.IsZero() && nextTime.After(now) {
					now = nextTime
				}

				nextTime = FindNext(job.schedules, now)
				until := time.Until(nextTime)

				if until <= 0 {
					until = time.Second // Ensure we wait at least a second before running the job again
				}

				if c.log != nil {
					c.log.Info("starting cron job", "job", job.Name, "spec", job.Specs, "next_run", nextTime, "remaining", until)
				}

				select {
				case <-ctx.Done():
					if c.log != nil {
						c.log.Info("stopping cron job", "job", job.Name)
					}

					return
				case <-time.After(until):
					if c.log != nil {
						c.log.Info("running cron job", "job", job.Name)
					}

					if err := job.Func(ctx); err != nil {
						if c.log != nil {
							c.log.Error("error running cron job", "job", job.Name, "error", err)
						}
					}
				}
			}
		}(job)
	}

	return nil
}

// Stop stops the cron job with cancel context and waits for all running jobs to finish.
func (c *cronJob) Stop() {
	c.m.Lock()
	defer c.m.Unlock()

	if !c.started {
		return
	}
	c.started = false

	c.cancel()
	c.wg.Wait()
}
