package main

import (
	"fmt"
	"log"
	"os"
)
import "github.com/urfave/cli/v2"

func main() {
	app := &cli.App{
		//Name:  "boom",
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
		},
		Action: func(c *cli.Context) error {
			fmt.Println("boom! I say!", c.String("app-id"), c.String("rest-api-key"))
			return nil
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
