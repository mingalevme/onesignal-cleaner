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
				Name: "ttl",
				Usage: "TTL in seconds, default 15552000 (3600*24*30*6)",
				EnvVars: []string{"ONESIGNAL_CLEANER_TTL"},
				Required: false,
			},
			&cli.BoolFlag{
				Name: "debug",
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
			if c.Int("ttl") > 0 {
				cleaner.TTL = c.Int("ttl")
			}
			logger.WithField("app-id", cleaner.OneSignalClient.AppId).
				WithField("ttl", cleaner.TTL).
				Infof("OneSignal cleaning is starting ...")
			err := cleaner.Clean()
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
