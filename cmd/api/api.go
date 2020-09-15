package main

import (
	"github.com/creekorful/quotes-api/internal/api"
	"github.com/labstack/gommon/log"
	"os"
)

func main() {
	a, err := api.NewAPI(os.Getenv("MONGO_URI"))
	if err != nil {
		log.Fatalf("unable to start API: %s", err)
	}

	log.Fatal(a.Start(":8080"))
}
