package main

import (
	"os"
	"time"

	"github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"

	"radiorecorder/config"
	"radiorecorder/stream"
)

func init() {
	// set moscow location for cron and for getting time.Now in Moscow tz
	if err := os.Setenv("TZ", "Europe/Moscow"); err != nil {
		log.Fatalf("Fail set TZ in env, err='%+v'", err)
	}
}

func main() {
	c := cron.New()
	// At 10:59 on Tuesday.
	if _, err := c.AddFunc("59 10 * * 2", downloadSerpNasheRadio); err != nil {
		log.Fatalf("Fail adding cron job, err='%+v'", err)
	}
	// At 10:59 on Thursday.
	if _, err := c.AddFunc("59 10 * * 4", downloadSerpNasheRadio); err != nil {
		log.Fatalf("Fail adding cron job, err='%+v'", err)
	}
	log.Info("Start cron")
	c.Run()
}

func downloadSerpNasheRadio() {
	duration := time.Hour + 10*time.Minute
	stream.NewDownloader(config.SerpNasheRadioUrl, "serp_nashe_radio_").Download(duration)
}
