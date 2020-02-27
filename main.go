package main

import (
	"os"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"

	"radiorecorder/config"
	"radiorecorder/stream"
)

func init() {
	log.SetOutput(os.Stdout)

	// set moscow location for cron and for getting time.Now in Moscow tz
	if err := os.Setenv("TZ", "Europe/Moscow"); err != nil {
		log.Fatalf("Fail set TZ in env, err='%+v'", err)
	}

	// set the SENTRY_DSN environment variable.
	if err := sentry.Init(sentry.ClientOptions{
		Dsn: config.SENTRY_DSN,
		// Enable printing of SDK debug messages.
		// Useful when getting started or trying to figure something out.
		Debug: true,
	}); err != nil {
		log.Fatalf("Fail init sentry err=%+v", err)
	}
}

func main() {
	c := cron.New()
	// At 11:00 on Tuesday.
	addTask(c, "0 11 * * 2", downloadDoObedaShow)
	// At 11:00 on Thursday.
	addTask(c, "0 11 * * 4", downloadDoObedaShow)
	// every hour
	addTask(c, "0 * * * *", ping)

	log.Info("Start cron")
	c.Run()
}

func addTask(c *cron.Cron, time string, f func()) {
	if _, err := c.AddFunc(time, f); err != nil {
		sentry.CaptureException(err)
		sentry.Flush(0)
		log.Fatalf("Fail adding cron job, err='%+v'", err)
	}
}

func ping() {
	log.Infof("Ping. time=%s", time.Now())
}


func downloadDoObedaShow() {
	downloader, err := stream.NewDownloader(
		config.SerpNasheRadioUrl,
		"do_obeda_show_nashe_radio",
		config.AudioDirectory,
	)
	if err != nil {
		log.Error("Fail getting downloader")
		sentry.CaptureException(err)
		return
	}

	duration := time.Hour
	if err := downloader.Download(duration); err != nil {
		sentry.CaptureException(err)
	}
}
