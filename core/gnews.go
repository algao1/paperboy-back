package core

import (
	"fmt"
	"log"
	"paperboy-back"
	"strings"
	"sync"
	"time"
)

func articleFilter(r *paperboy.Result) bool {
	return strings.Contains(r.Title, "| Letters")
}

func summarize(gs paperboy.GuardianService, res []*paperboy.Result,
	sch chan<- *paperboy.Summary, ech chan<- error) {
	var wg sync.WaitGroup

	for _, r := range res {
		// Blacklist articles based on filter.
		if articleFilter(r) {
			continue
		}

		wg.Add(1)
		go func(r *paperboy.Result) {
			defer wg.Done()

			summ, err := gs.ExtractOne(r)
			if err != nil {
				ech <- err
			}
			sch <- summ
		}(r)
	}

	wg.Wait()
	close(sch)
}

// GuardianNews returns a Tasker that will periodically fetch news from the Guardian API.
func GuardianNews(section string, hours int, ss paperboy.SummaryService,
	gs paperboy.GuardianService, tf paperboy.TaskerFactory) (paperboy.Tasker, error) {
	// Defines the task.
	task := func() error {
		start := time.Now()
		qparams := map[string]string{
			"section":     section,
			"type":        "article",
			"show-fields": "trailText,wordcount,bodyText",
			"show-tags":   "contributor",
			"show-blocks": "main",
			"page-size":   "50",
			"from-date":   time.Now().UTC().Add(time.Duration(hours) * time.Hour).Format("2006-01-02T15:04:05.999999"),
		}

		g, err := gs.Fetch(qparams)
		if err != nil {
			return fmt.Errorf("%q: %w", "could not fetch from Guardian", err)
		}

		// Create channels and waitGroup.
		errCh := make(chan error)
		sumCh := make(chan *paperboy.Summary)

		// Assign each summarization to a seperate goroutine.
		go summarize(gs, g.Response.Results, sumCh, errCh)

		// Receive summaries and error over channels.
		for {
			select {
			case s, ok := <-sumCh:
				// Wrap up the task on channel close.
				if !ok {
					log.Printf("[Guardian News - %s] summarized %d articles in %v",
						strings.Title(section),
						len(g.Response.Results),
						time.Since(start),
					)
					return nil
				}
				ss.Create(s)
			case err := <-errCh:
				close(errCh)
				return fmt.Errorf("%q: %w", "failed to summarize news", err)
			}
		}
	}

	// Configures and returns a Tasker.
	name := fmt.Sprintf("Guardian %s", strings.Title(section))
	conf := paperboy.TaskConfig{Name: name, Period: 1 * time.Hour, RecoverPeriod: 5 * time.Minute}
	guardianNews, err := tf.CreateTasker(conf, task)
	if err != nil {
		return guardianNews, fmt.Errorf("%q: %w", "could not create Guardian tasker", err)
	}

	return guardianNews, nil
}
