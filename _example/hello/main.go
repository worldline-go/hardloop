package main

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/worldline-go/hardloop"
)

func MyFunction(ctx context.Context, wg *sync.WaitGroup) error {
	// wg is used to wait for the function to finish.
	// if you open a new goroutine, you must add 1 to wg.Add()
	// it is help to finish the function before the start again.

	// do something
	log.Println("Hello hardloop!")
	log.Println("Waiting 12 secs...")

	timer := time.NewTimer(12 * time.Second)

	select {
	case <-timer.C:
		log.Println("Done!")
	case <-ctx.Done():
		log.Println("Canceled!")
		// stop timer
		timer.Stop()
		select {
		case <-timer.C:
		default:
		}
	}

	return nil
}

type myLog struct{}

func (myLog) Error(msg string, keysAndValues ...interface{}) {
	log.Printf(msg, keysAndValues...)
}
func (myLog) Info(msg string, keysAndValues ...interface{}) {
	log.Printf(msg, keysAndValues...)
}
func (myLog) Debug(msg string, keysAndValues ...interface{}) {
	log.Printf(msg, keysAndValues...)
}
func (myLog) Warn(msg string, keysAndValues ...interface{}) {
	log.Printf(msg, keysAndValues...)
}

func main() {
	// Set start cron specs.
	startSpecs := []string{
		// start at 7:00 in Monday, Tuesday, Wednesday, Thursday, Friday
		"CRON_TZ=Europe/Amsterdam 0 7 * * 1,2,3,4,5",
		// start at 9:00 in Saturday
		"0 9 * * 6",
	}
	// Set stop cron specs.
	stopSpecs := []string{
		// stop at 17:00 in Monday, Tuesday, Wednesday, Thursday, Friday
		"0 17 * * 1,2,3,4,5",
		// stop at 13:00 in Saturday
		"0 13 * * 6",
	}

	// Create a new schedule.
	myFunctionLoop, err := hardloop.NewLoop(startSpecs, stopSpecs, MyFunction)
	if err != nil {
		// wrong cron specs
		log.Fatal(err)
	}

	myFunctionLoop.SetLogger(myLog{})

	// run forever in goroutine (or until the function returns ErrLoopExited)
	myFunctionLoop.RunWait(context.Background())
}
