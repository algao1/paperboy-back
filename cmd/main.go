package main

import (
	"log"
	"os"
	"paperboy-back/chi"
	"paperboy-back/core"
	"paperboy-back/mongo"
	"paperboy-back/news/guardian"
	"paperboy-back/redis"
	"paperboy-back/tasker"
	"strconv"

	"github.com/joho/godotenv"
)

func main() {
	// Loads configurations.
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Initialize services, factories, and handlers.
	ss, err := mongo.Open(
		os.Getenv("MONGO_URI"),
		os.Getenv("MONGO_KEY"),
		os.Getenv("MONGO_DB"),
		os.Getenv("MONGO_COL"),
	)
	if err != nil {
		log.Fatal(err)
	}

	cdb, err := strconv.Atoi(os.Getenv("CACHE_DB"))
	if err != nil {
		log.Fatal(err)
	}
	ssc := redis.NewSummaryCache(os.Getenv("CACHE_URL"), os.Getenv("CACHE_PORT"), os.Getenv("CACHE_PASS"), cdb, ss)

	gs := guardian.Create(os.Getenv("GUARDIAN_KEY"))
	tf := &tasker.Factory{}
	h := chi.Init(ssc)

	// Dependency injection.
	serv := core.Server{
		SummaryService:  ssc,
		GuardianService: gs,
		TaskerFactory:   tf,
		Handler:         h,
	}

	// Run the server.
	port, err := strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("running on port:%d", port)

	if err = serv.Run(port); err != nil {
		log.Fatal(err)
	}
}
