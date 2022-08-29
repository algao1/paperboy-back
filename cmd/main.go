package main

import (
	"log"
	"os"
	"paperboy-back/chi"
	"paperboy-back/core"
	"paperboy-back/mongo"
	"paperboy-back/news/guardian"
	"paperboy-back/tasker"
)

func main() {
	// Initialize services, factories, and handlers.
	ss, err := mongo.Open(
		os.Getenv("MONGO_URI"),
		os.Getenv("MONGO_KEY"),
	)
	if err != nil {
		log.Fatal(err)
	}

	// TODO: should really fix this in the future.

	// cdb, err := strconv.Atoi(os.Getenv("CACHE_DB"))
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// ssc := redis.NewSummaryCache(
	// 	os.Getenv("CACHE_URL"),
	// 	os.Getenv("CACHE_PORT"),
	// 	os.Getenv("CACHE_PASS"),
	// 	cdb, ss)

	gs := guardian.Create(os.Getenv("GUARDIAN_KEY"))
	tf := &tasker.Factory{}
	h := chi.Init(ss)

	// Dependency injection.
	serv := core.Server{
		SummaryService:  ss,
		GuardianService: gs,
		TaskerFactory:   tf,
		Handler:         h,
	}

	log.Println("running on port 8080")

	if err = serv.Run(8080); err != nil {
		log.Fatal(err)
	}
}
