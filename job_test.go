package hardloop_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/worldline-go/hardloop"
)

func TestJob(t *testing.T) {
	t.Skip("skipping time-based test")

	// Define a simple cron job function
	jobFunc := func(ctx context.Context) error {
		fmt.Println("Cron job executed")
		return nil
	}

	// Create a new cron job
	cronJob, err := hardloop.NewCron(hardloop.Cron{
		Name:  "TestJob",
		Func:  jobFunc,
		Specs: []string{"* * * * *"}, // Every minute
	})
	if err != nil {
		t.Fatalf("Failed to create cron job: %v", err)
	}

	// Start the cron job
	if err := cronJob.Start(t.Context()); err != nil {
		t.Fatalf("Failed to start cron job: %v", err)
	}

	<-t.Context().Done()
}
