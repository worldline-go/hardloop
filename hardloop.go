package hardloop

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

var (
	// GapDurationStart to start the function should be at least 1 second bigger than gap duration.
	GapDurationStart time.Duration = 1 * time.Second //nolint:revive // more readable
	// GapDurationStop to stop the function.
	GapDurationStop time.Duration = 0 //nolint:revive // more readable

	// ErrCloseLoop is returned when the loop should be closed.
	ErrCloseLoop  = errors.New("close loop")
	errTimeNotSet = errors.New("timeless schedule")
)

type Loop struct {
	startSchedules    []Schedule
	stopSchedules     []Schedule
	isLoopRunning     bool
	isFunctionRunning bool
	fn                func(ctx context.Context) error
	mx                sync.RWMutex
	cancelFn          context.CancelFunc
	cancelLoop        context.CancelFunc
	exited            chan struct{}
	startDuration     chan *time.Duration
	stopDuration      chan *time.Duration
	log               Logger
}

// NewLoop returns a new Loop with the given start and end cron specs and function.
//   - Standard crontab specs, e.g. "* * * * ?"
//   - Descriptors, e.g. "@midnight", "@every 1h30m"
func NewLoop(startSpec, endSpec []string, fn func(ctx context.Context) error) (*Loop, error) {
	startSchedules := make([]Schedule, 0, len(startSpec))
	stopSchedules := make([]Schedule, 0, len(endSpec))

	for _, spec := range startSpec {
		startSchedule, err := ParseStandard(spec)
		if err != nil {
			return nil, err
		}

		startSchedules = append(startSchedules, startSchedule)
	}

	for _, spec := range endSpec {
		stopSchedule, err := ParseStandard(spec)
		if err != nil {
			return nil, err
		}

		stopSchedules = append(stopSchedules, stopSchedule)
	}

	return &Loop{
		startSchedules:    startSchedules,
		stopSchedules:     stopSchedules,
		isLoopRunning:     false,
		isFunctionRunning: false,
		fn:                fn,
		exited:            make(chan struct{}, 1),
		startDuration:     make(chan *time.Duration, 1),
		stopDuration:      make(chan *time.Duration, 1),
		log:               slog.Default(),
	}, nil
}

// SetLogger sets the logger for the loop.
//   - If not set, it uses the default slog logger.
//   - Set to nil to disable logging.
func (l *Loop) SetLogger(log Logger) {
	l.log = log
}

// ChangeStartSchedules sets the start cron specs.
// Not effects immediately!
func (l *Loop) ChangeStartSchedules(startSpecs []string) error {
	startSchedules := make([]Schedule, 0, len(startSpecs))

	for _, spec := range startSpecs {
		startSchedule, err := ParseStandard(spec)
		if err != nil {
			return err
		}

		startSchedules = append(startSchedules, startSchedule)
	}

	l.startSchedules = startSchedules

	return nil
}

// ChangeStopSchedule sets the end cron specs.
// Not effects immediately!
func (l *Loop) ChangeStopSchedules(stopSpecs []string) error {
	stopSchedules := make([]Schedule, 0, len(stopSpecs))

	for _, spec := range stopSpecs {
		stopSchedule, err := ParseStandard(spec)
		if err != nil {
			return err
		}

		stopSchedules = append(stopSchedules, stopSchedule)
	}

	l.stopSchedules = stopSchedules

	return nil
}

// IsLoopRunning returns true if the loop is running.
func (l *Loop) IsLoopRunning() bool {
	l.mx.RLock()
	defer l.mx.RUnlock()

	return l.isLoopRunning
}

// IsFunctionRunning returns true if the function is running.
func (l *Loop) IsFunctionRunning() bool {
	l.mx.RLock()
	defer l.mx.RUnlock()

	return l.isFunctionRunning
}

// RunWait starts the loop and wait to exit with ErrLoopExited.
func (l *Loop) RunWait(ctx context.Context) {
	wg := &sync.WaitGroup{}
	l.Run(ctx, wg)
	wg.Wait()
}

// Run starts the loop.
func (l *Loop) Run(ctx context.Context, wg *sync.WaitGroup) {
	if l.IsLoopRunning() {
		return
	}

	var ctxLoop context.Context
	ctxLoop, l.cancelLoop = context.WithCancel(ctx)

	// listen function exit
	wg.Add(1)
	go func() {
		defer wg.Done()

		for {
			select {
			case <-ctxLoop.Done():
				return
			case <-l.exited:
				now := time.Now().Add(GapDurationStart)
				// check it can run in now
				stopTime, _ := l.getStopTime(now)
				if stopTime != nil {
					l.runFunction(ctxLoop, wg)

					continue
				}

				now = time.Now()
				// check next time to start again
				startTime, _ := l.getStartTime(now)
				if startTime == nil {
					// disable next start time
					if l.log != nil {
						l.log.Info("Next start time disabled")
					}
					l.startDuration <- nil

					continue
				}

				// set next start time
				if l.log != nil {
					l.log.Info(fmt.Sprintf("Next start time: [%s]", startTime))
				}
				duration := startTime.Sub(now)
				l.startDuration <- &duration
			}
		}
	}()

	// listen start time
	wg.Add(1)
	go func() {
		defer wg.Done()

		var chStartDuration <-chan time.Time
		var startTimer *time.Timer

		for {
			select {
			case <-ctxLoop.Done():
				return
			case startDuration := <-l.startDuration:
				if startDuration == nil {
					// disable
					chStartDuration = nil

					// run now
					l.runFunction(ctxLoop, wg)

					continue
				}

				// set next start time
				startTimerChange := time.NewTimer(*startDuration)
				chStartDuration = startTimerChange.C

				if startTimer != nil {
					startTimer.Stop()
					select {
					case <-startTimer.C:
					default:
					}
				}
				startTimer = startTimerChange
			case <-chStartDuration:
				// run function
				l.runFunction(ctxLoop, wg)
			}
		}
	}()

	// listen stop time
	wg.Add(1)
	go func() {
		defer wg.Done()

		var chStopDuration <-chan time.Time
		var stopTimer *time.Timer

		for {
			select {
			case <-ctxLoop.Done():
				return
			case stopDuration := <-l.stopDuration:
				if stopDuration == nil {
					// disable
					chStopDuration = nil

					continue
				}

				// set next stop time and clear the previous one
				stopTimerChange := time.NewTimer(*stopDuration)
				chStopDuration = stopTimerChange.C

				if stopTimer != nil {
					stopTimer.Stop()
					select {
					case <-stopTimer.C:
					default:
					}
				}
				stopTimer = stopTimerChange
			case <-chStopDuration:
				// time to stop function
				l.stopFunction()
			}
		}
	}()

	// first initialize
	l.initializeTime(ctxLoop, wg)
}

func (l *Loop) runFunction(ctx context.Context, wg *sync.WaitGroup) {
	l.mx.Lock()
	defer l.mx.Unlock()

	if l.isFunctionRunning {
		return
	}

	l.isFunctionRunning = true

	wg.Add(1)
	go func() {
		defer wg.Done()
		var ctxInFunc context.Context
		ctxInFunc, l.cancelFn = context.WithCancel(ctx)
		err := l.fn(ctxInFunc)

		// set running to false
		l.mx.Lock()
		defer l.mx.Unlock()
		l.isFunctionRunning = false

		if errors.Is(err, ErrCloseLoop) {
			l.cancelLoop()

			return
		}
		// trigger exited
		l.exited <- struct{}{}
	}()

	// set next stop time
	now := time.Now().Add(GapDurationStart)
	stopTime, _ := l.getStopTime(now)
	if stopTime == nil {
		// disable next stop time
		if l.log != nil {
			l.log.Info("Next stop time disabled")
		}

		l.stopDuration <- nil

		return
	}

	if l.log != nil {
		l.log.Info(fmt.Sprintf("Next stop time: [%s]", stopTime))
	}

	stopDuration := stopTime.Sub(now)
	l.stopDuration <- &stopDuration
}

func (l *Loop) stopFunction() {
	l.mx.Lock()
	defer l.mx.Unlock()

	// if function is not running, trigger exited to get the next start time
	if !l.isFunctionRunning {
		// trigger exited
		l.exited <- struct{}{}

		return
	}

	l.isFunctionRunning = false

	l.cancelFn()
}

func (l *Loop) initializeTime(ctx context.Context, wg *sync.WaitGroup) {
	v, _ := l.getStopTime(time.Now().Add(GapDurationStart))
	if v != nil {
		// function should run now
		l.runFunction(ctx, wg)

		return
	}

	// set next start time
	l.exited <- struct{}{}
}

// getStartTime if return nil, start now.
func (l *Loop) getStartTime(now time.Time) (*time.Time, error) {
	nextStart := FindNext(l.startSchedules, now)

	if nextStart.IsZero() {
		return nil, errTimeNotSet
	}

	return &nextStart, nil
}

// getStopTime if return nil, stop now.
func (l *Loop) getStopTime(now time.Time) (*time.Time, error) {
	prevStop := FindPrev(l.stopSchedules, now)

	if prevStop.IsZero() {
		// stop the loop
		return nil, errTimeNotSet
	}

	prevStart := FindPrev(l.startSchedules, now)

	// if prevStop is after prevStart, then we should stop the loop
	if !prevStart.IsZero() && prevStop.After(prevStart) {
		// stop the loop
		return nil, nil
	}

	nextStop := FindNext(l.stopSchedules, now)

	if nextStop.IsZero() {
		// stop the loop
		return nil, errTimeNotSet
	}

	return &nextStop, nil
}
