# hardloop

[![License](https://img.shields.io/github/license/worldline-go/hardloop?color=red&style=flat-square)](https://raw.githubusercontent.com/worldline-go/hardloop/main/LICENSE)
[![Coverage](https://img.shields.io/sonar/coverage/worldline-go_hardloop?logo=sonarcloud&server=https%3A%2F%2Fsonarcloud.io&style=flat-square)](https://sonarcloud.io/summary/overall?id=worldline-go_hardloop)
[![GitHub Workflow Status](https://img.shields.io/github/actions/workflow/status/worldline-go/hardloop/test.yml?branch=main&logo=github&style=flat-square&label=ci)](https://github.com/worldline-go/hardloop/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/worldline-go/hardloop?style=flat-square)](https://goreportcard.com/report/github.com/worldline-go/hardloop)
[![Go PKG](https://raw.githubusercontent.com/worldline-go/guide/main/badge/custom/reference.svg)](https://pkg.go.dev/github.com/worldline-go/hardloop)

Hardloop is a cron time-based function runner.

Set start and end times as cron specs, and give function to run between them.

```shell
go get github.com/worldline-go/hardloop
```

![hardloop](./_assets/hardloop.svg)

## Usage

Check the https://crontab.guru/ to explain about cron specs.

> Hardloop different works than _crontab.guru_ in weekdays and day of month selection. We use __and__ operation but that site use __or__ operation when used both of them.

You can give as much as you want start, stop times.  
If stop time is not given, it will run forever.  
If just stop time given, it will restart in the stop times.

Use Timezone to set the timezone: `CRON_TZ=Europe/Istanbul 0 7 * * 1,2,3,4,5` 

Default timezone is system timezone.

> Some times doens't exist in some timezones.
> For example, `CRON_TZ=Europe/Amsterdam 30 2 26 3 *` doesn't exist due to in that time 02:00 -> 03:00 DST 1 hour adding. It will be make some problems so don't use non-exist times.

```go
// Set start cron specs.
startSpecs := []string{
    // start at 7:00 in Monday, Tuesday, Wednesday, Thursday, Friday
    "0 7 * * 1,2,3,4,5",
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

// run forever in goroutine (or until the function returns ErrLoopExited)
myFunctionLoop.RunWait(ctx)
```

For simple jobs, you can use `hardloop.NewCron` to create a cron job.

```go
// Create a new cron job.
myCronJob, err := hardloop.NewCron(hardloop.Cron{
	Name:  "MyCronJob",
	Func:  MyFunction,
	Specs: []string{"0 7 * * 1-5"}, // Every weekday at 7 AM
})
// ... handle error

myCronJob.Start(ctx) // background run the cron job
// myCronJob.Stop() // to stop the cron job
```

### Set Logger

Implement Logger interface and set to the loop.

```go
myFunctionLoop.SetLogger(myLog{})
```

## Development

Test code

```sh
make test
# make coverage html-wsl
```
