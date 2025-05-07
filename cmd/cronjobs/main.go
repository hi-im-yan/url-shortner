package main

import (
	"log"
	"url-shortner/internal/database"

	"github.com/robfig/cron/v3"
)

func main() {

	log.Println("[cronjobs:main] Running cronjob")
	c := cron.New()

	db := database.New()

	// Running every minute
	c.AddFunc("*/1 * * * *", func() {
		db.DeleteExpiredLinks()
	})

	c.Start()

	// This keeps the program running
	select {}
}
