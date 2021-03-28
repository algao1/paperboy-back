package chi

import (
	"encoding/json"
	"log"
	"net/http"
	"paperboy-back"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// Handler is an implementation of http.Handler.
type Handler struct {
	chi http.Handler
}

var _ http.Handler = (*Handler)(nil)

// Init configures and returns a chi router.
func Init(ss paperboy.SummaryService) *Handler {
	r := chi.NewRouter()

	// Middleware.
	r.Use(middleware.Logger)
	r.Use(middleware.Timeout(60 * time.Second))

	// RESTy routes for 'summaries' resource.
	r.Get("/api/summary", apiGetSummary(ss))
	r.Get("/api/summaries", apiSearchSummaries(ss))
	r.Get("/api/summaries/{section}", apiGetSummaries(ss))

	return &Handler{chi: r}
}

// ServeHTTP is a wrapper function for chi.ServeHTTP.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.chi.ServeHTTP(w, r)
}

// Closure to bind SummaryService to the HandlerFunc in order to search summaries.
func apiSearchSummaries(ss paperboy.SummaryService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Obtain the query parameter 'q'.
		query := r.URL.Query().Get("q")

		summaries, err := ss.Search(query)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		log.Printf("[%s] found %d summaries", r.URL, len(summaries))

		// Marshals slice of summaries into []bytes.
		js, err := json.Marshal(summaries)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Sets and writes content-type of 'application/json'.
		w.Header().Set("Content-Type", "application/json")
		w.Write(js)
	}
}

// Closure to bind SummaryService to the HandlerFunc in order to serve summaries.
func apiGetSummaries(ss paperboy.SummaryService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		section := chi.URLParam(r, "section")

		// Obtain the query parameter 'id'.
		id := r.URL.Query().Get("id")

		// Obtain the query parameter 'size'.
		ssize := r.URL.Query().Get("size")
		size, err := strconv.Atoi(ssize)
		if err != nil {
			log.Printf("[%s] query param 'recent=%s' is invalid\n", r.URL, ssize)
			size = 10
		}

		// Fetch summaries using SummaryService.
		summaries, last, err := ss.Summaries(section, id, size)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		log.Printf("[%s] fetched summaries\n", r.URL)

		// Marshals slice of summaries into []bytes.
		js, err := json.Marshal(paperboy.SummariesResponse{LastID: last, Summaries: summaries})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Sets and writes content-type of 'application/json'.
		w.Header().Set("Content-Type", "application/json")
		w.Write(js)
	}
}

func apiGetSummary(ss paperboy.SummaryService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Obtain the query parameter 'id'.
		id := r.URL.Query().Get("id")

		summary, err := ss.Summary(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		js, err := json.Marshal(summary)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Sets and writes content-type of 'application/json'.
		w.Header().Set("Content-Type", "application/json")
		w.Write(js)
	}
}
