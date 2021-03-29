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
	// Start the tasks.
	gworld, err := GuardianNews("world", -24, s.SummaryService, s.GuardianService, s.TaskerFactory)
	if err != nil {
		return fmt.Errorf("%q: %w", "unable to start server", err)
	}
	gworld.Start()

	genv, err := GuardianNews("environment", -24, s.SummaryService, s.GuardianService, s.TaskerFactory)
	if err != nil {
		return fmt.Errorf("%q: %w", "unable to start server", err)
	}
	genv.Start()

	gtech, err := GuardianNews("technology", -24, s.SummaryService, s.GuardianService, s.TaskerFactory)
	if err != nil {
		return fmt.Errorf("%q: %w", "unable to start server", err)
	}
	gtech.Start()

	gsci, err := GuardianNews("technology", -24, s.SummaryService, s.GuardianService, s.TaskerFactory)
	if err != nil {
		return fmt.Errorf("%q: %w", "unable to start server", err)
	}
	gsci.Start()

	http.ListenAndServe(fmt.Sprintf(":%v", port), s.Handler)
	return nil
}
