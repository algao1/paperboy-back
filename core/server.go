package core

import (
	"fmt"
	"net/http"
	"paperboy-back"
)

// Server contains all the dependencies required for the application.
type Server struct {
	SummaryService  paperboy.SummaryService
	GuardianService paperboy.GuardianService
	TaskerFactory   paperboy.TaskerFactory
	Handler         http.Handler
}

// Run starts the server at the designated port.
func (s *Server) Run(port int) error {
	// Create channel for communicating summaries.
	summaries := make(chan *paperboy.Summary)

	// Start the tasks.
	gworld, err := GuardianNews("world", summaries, s.GuardianService, s.TaskerFactory)
	if err != nil {
		return fmt.Errorf("%q: %w", "could not start server", err)
	}
	gworld.Start()

	go func() {
		for sum := range summaries {
			s.SummaryService.Create(sum)
		}
	}()

	http.ListenAndServe(fmt.Sprintf(":%v", port), s.Handler)
	return nil
}
