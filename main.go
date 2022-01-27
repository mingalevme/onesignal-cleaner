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
				Name: "inactivity-threshold",
				Usage: "Max time in seconds player is considered alive, default is 6 months (15552000s)",
				EnvVars: []string{"ONESIGNAL_CLEANER_INACTIVITY_THRESHOLD"},
				Required: false,
			},
			&cli.IntFlag{
				Name: "connection-timeout",
				Usage: "Max time in seconds to wait players data resource is ready",
				EnvVars: []string{"ONESIGNAL_CLEANER_CONNECTION_TIMEOUT"},
				Required: false,
			},
			&cli.StringFlag{
				Name: "tmp-dir",
				Usage: "Dir to download players data",
				EnvVars: []string{"ONESIGNAL_CLEANER_TMP_DIR"},
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
			if c.Int("inactivity-threshold") > 0 {
				cleaner.TTL = c.Int("inactivity-threshold")
			}
			if c.String("tmp-dir") != "" {
				cleaner.TmpDir = c.String("tmp-dir")
			}
			if c.Int("connection-timeout") > 0 {
				cleaner.Downloader.ReadinessTimeout = c.Int("connection-timeout")
			}
			if c.Int("concurrency") > 0 {
				cleaner.Concurrency = c.Int("concurrency")
			}
			logger.WithField("app-id", cleaner.OneSignalClient.AppId).
				WithField("inactivity-threshold", cleaner.TTL).
				WithField("concurrency", cleaner.Concurrency).
				WithField("connection-timeout", cleaner.Downloader.ReadinessTimeout).
				WithField("tmp-dir", cleaner.TmpDir).
				Infof("OneSignal cleaning is starting ...")
			err := cleaner.Clean(c.String("data-file"))
			if err != nil {
				logger.WithField("app-id", cleaner.OneSignalClient.AppId).
					WithField("ttl", cleaner.TTL).
					WithError(err).
					Errorf("Error while OneSignal cleaning")
			} else {
				logger.WithField("app-id", cleaner.OneSignalClient.AppId).
					WithField("ttl", cleaner.TTL).
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
