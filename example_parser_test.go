package hardloop_test

import (
	"fmt"
	"log"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/worldline-go/hardloop"
)

func Example_parser() {
	schedule := "0 7 * * *"

	p := hardloop.Parser{ParseFn: cron.ParseStandard}

	s, err := p.Parse2(schedule)
	if err != nil {
		log.Fatal(err)
	}

	now := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)

	fmt.Println("Next 5 times:")
	tmpNow := now
	for i := 0; i < 5; i++ {
		tmpNow = s.Next(tmpNow)
		fmt.Println(tmpNow)
	}

	tmpNow = now
	fmt.Println("Prev 5 times:")
	for i := 0; i < 5; i++ {
		tmpNow = s.Prev(tmpNow)
		fmt.Println(tmpNow)
	}

	// Output:
	// Next 5 times:
	// 2023-01-01 07:00:00 +0000 UTC
	// 2023-01-02 07:00:00 +0000 UTC
	// 2023-01-03 07:00:00 +0000 UTC
	// 2023-01-04 07:00:00 +0000 UTC
	// 2023-01-05 07:00:00 +0000 UTC
	// Prev 5 times:
	// 2022-12-31 07:00:00 +0000 UTC
	// 2022-12-30 07:00:00 +0000 UTC
	// 2022-12-29 07:00:00 +0000 UTC
	// 2022-12-28 07:00:00 +0000 UTC
	// 2022-12-27 07:00:00 +0000 UTC
}
