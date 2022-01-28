package main

import (
	"github.com/mingalevme/gologger"
	"log"
	"os"
)
import "github.com/urfave/cli/v2"

func main() {
	app := &cli.App{
		Name:  "clean",
		Usage: "make an explosive entrance",
		Flags: []cli.Flag {
			&cli.StringFlag{
				Name: "app-id",
				Usage: "OneSignal App ID",
				EnvVars: []string{"ONESIGNAL_CLEANER_APP_ID"},
				Required: true,
			},
			&cli.StringFlag{
				Name: "rest-api-key",
				Usage: "Rest API Key",
				EnvVars: []string{"ONESIGNAL_CLEANER_REST_API_KEY"},
				Required: true,
			},
			&cli.IntFlag{
				Name: "inactive-for",
				Usage: "Max time in seconds player is considered active, default is 1 year",
				EnvVars: []string{"ONESIGNAL_CLEANER_INACTIVITY_THRESHOLD"},
				Value: 86400*365,
				Required: false,
			},
			&cli.IntFlag{
				Name: "readiness-timeout",
				Usage: "Max time in seconds to wait players data resource is ready",
				EnvVars: []string{"ONESIGNAL_CLEANER_CONNECTION_TIMEOUT"},
				Value: 600,
				Required: false,
			},
			&cli.StringFlag{
				Name: "tmp-dir",
				Usage: "Dir to download players data, default is operation system's temporary directory",
				EnvVars: []string{"ONESIGNAL_CLEANER_TMP_DIR"},
				Value: os.TempDir(),
				Required: false,
			},
			&cli.IntFlag{
				Name: "concurrency",
				Usage: "Max number of concurrent requests",
				EnvVars: []string{"ONESIGNAL_CLEANER_CONCURRENCY"},
				Value: 5,
				Required: false,
			},
			&cli.StringFlag{
				Name: "data-file",
				Usage: "Read data from a local file (*.csv.gz) instead of requesting one from OneSignal",
				EnvVars: []string{"ONESIGNAL_CLEANER_DATA_FILE"},
				Required: false,
			},
			&cli.BoolFlag{
				Name: "download-only",
				Usage: "Download data file only without handling it",
				EnvVars: []string{"ONESIGNAL_CLEANER_DOWNLOAD_ONLY"},
				Value: false,
				Required: false,
			},
			&cli.BoolFlag{
				Name: "debug",
				Usage: "Sets logging level to debug",
				EnvVars: []string{"ONESIGNAL_CLEANER_DEBUG"},
				Required: false,
			},
		},
		Action: func(c *cli.Context) error {
			lvl := gologger.LevelInfo
			if c.Bool("debug") {
				lvl = gologger.LevelDebug
			}
			logger := gologger.NewStdoutLogger(lvl)
			cleaner := NewCleaner(c.String("app-id"), c.String("rest-api-key"), logger)
			cleaner.Logger = logger
			if c.Int("inactive-for") > 0 {
				cleaner.InactiveFor = c.Int("inactive-for")
			}
			if c.String("tmp-dir") != "" {
				cleaner.TmpDir = c.String("tmp-dir")
			}
			if c.Int("readiness-timeout") > 0 {
				cleaner.Downloader.ReadinessTimeout = c.Int("readiness-timeout")
			}
			if c.Int("concurrency") > 0 {
				cleaner.Concurrency = c.Int("concurrency")
			}
			if c.Bool("download-only") {
				logger.Infof("Starting in \"download-only\"-mode")
				cleaner.DownloadOnly = true
			}
			logger.WithField("app-id", cleaner.OneSignalClient.AppId).
				WithField("inactive-for", cleaner.InactiveFor).
				WithField("concurrency", cleaner.Concurrency).
				WithField("readiness-timeout", cleaner.Downloader.ReadinessTimeout).
				WithField("tmp-dir", cleaner.TmpDir).
				Infof("OneSignal cleaning is starting ...")
			err := cleaner.Clean(c.String("data-file"))
			if err != nil {
				logger.WithField("app-id", cleaner.OneSignalClient.AppId).
					WithField("inactive-for", cleaner.InactiveFor).
					WithError(err).
					Errorf("Error while OneSignal cleaning")
			} else {
				logger.WithField("app-id", cleaner.OneSignalClient.AppId).
					WithField("inactive-for", cleaner.InactiveFor).
					Infof("OneSignal cleaning has been finished successfully")
			}
			return nil
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
