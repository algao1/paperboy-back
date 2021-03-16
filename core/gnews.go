package core

import (
	"fmt"
	"log"
	"paperboy-back"
	"strings"
	"sync"
	"time"
)

// GuardianNews returns a Tasker that will periodically fetch news from the Guardian API.
func GuardianNews(section string, sch chan<- *paperboy.Summary,
	gs paperboy.GuardianService, tf paperboy.TaskerFactory) (paperboy.Tasker, error) {
	// Define the task.
	task := func(sch chan<- *paperboy.Summary, gs paperboy.GuardianService) error {
		start := time.Now()
		qparams := map[string]string{
			"section":     section,
			"type":        "article",
			"show-fields": "trailText,wordcount,bodyText",
			"show-tags":   "contributor",
			"show-blocks": "main",
			"from-date":   time.Now().UTC().Add(-1 * time.Hour).Format("2006-01-02T15:04:05.999999"),
		}

		g, err := gs.Fetch(qparams)
		if err != nil {
			return fmt.Errorf("%q: %w", "could not fetch from Guardian", err)
		}

		// Create channels and waitGroup.
		errCh := make(chan error)
		waitCh := make(chan struct{})
		var wg sync.WaitGroup

		// Assign each summarization to a seperate goroutine.
		go func() {
			for _, res := range g.Response.Results {
				wg.Add(1)
				go func(res *paperboy.Result) {
					defer wg.Done()
					summ, err := gs.ExtractOne(res)
					if err != nil {
						errCh <- err
					}
					sch <- summ
				}(res)
			}

			// Close the channel when finished.
			wg.Wait()
			close(waitCh)
		}()

		// Receive over channel with select.
		select {
		case <-waitCh:
			log.Printf("[Guardian News - %s] summarized %d articles in %v",
				strings.Title(section),
				len(g.Response.Results),
				time.Since(start))
			return nil
		case err := <-errCh:
			return fmt.Errorf("%q: %w", "failed to summarize news", err)
		}
	}

	// Configure and return Tasker.
	conf := paperboy.TaskConfig{Name: "Guardian World", Period: 1 * time.Hour, RecoverPeriod: 5 * time.Minute}
	guardianNews, err := tf.CreateTasker(conf, task, sch, gs)
	if err != nil {
		return guardianNews, fmt.Errorf("%q: %w", "could not create Guardian tasker", err)
	}
	return guardianNews, nil
}
