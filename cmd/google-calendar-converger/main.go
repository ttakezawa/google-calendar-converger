package main

import (
	"flag"
	"log"
	"os"
	"time"

	"github.com/ttakezawa/google-calendar-converger/pkg/converger"
	"github.com/ttakezawa/google-calendar-converger/pkg/event"
)

func main() {
	var (
		init              = flag.Bool("init", false, "init oauth token")
		titlePrefixFilter = flag.String("title-prefix-filter", "", "title prefix filter")
	)
	flag.Parse()

	cv := converger.New()
	cv.Init()

	if *init {
		return
	}

	if titlePrefixFilter == nil || *titlePrefixFilter == "" {
		log.Fatalf("-title-prefix-filter is not specified.")
	}

	events, err := event.Read(os.Stdin)
	if err != nil {
		log.Fatalf("%w", err)
	}

	cv.Run(time.Now(), *titlePrefixFilter, events)
}
