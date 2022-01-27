package main

import (
	"fmt"
	"github.com/mingalevme/gologger"
	"github.com/pkg/errors"
	"io"
	"os"
	"time"
)

type PlayerData map[string]string
type Player struct {
	Id         string
	LastActive time.Time
}

type Cleaner struct {
	OneSignalClient    *OneSignalClient
	Downloader         *Downloader
	GzCsvReaderFactory func(filename string) (*GzCsvReader, error)
	Logger             gologger.Logger
	TTL                int
	ConnectionTimeout  int
	TmpDir             string
	Concurrency        int
	Now                Nower
}

func NewCleaner(appId string, restApiKey string, logger gologger.Logger) *Cleaner {
	osc := NewOneSignalClient(appId, restApiKey)
	osc.Logger = logger
	d := NewDownloader()
	d.Logger = logger
	return &Cleaner{
		OneSignalClient: osc,
		Downloader:      d,
		GzCsvReaderFactory: func(filename string) (*GzCsvReader, error) {
			return NewGzCsvReader(filename)
		},
		TTL:    86400 * 30 * 6,
		TmpDir: os.TempDir(),
		Concurrency: 1,
		Now:    Now,
		Logger: logger,
	}
}

func (c *Cleaner) Clean() error {
	dataUrl, err := c.OneSignalClient.GetExportUrl()
	if err != nil {
		return errors.Wrap(err, "error while getting export url")
	}
	c.Logger.Infof("Export url has been fetched: %s", dataUrl)
	fileName := c.getDestFileName()
	f, err := os.Create(fileName)
	if err != nil {
		return errors.Wrap(err, "error while creating a temporary file")
	}
	defer func(f *os.File) {
		_ = f.Close()
	}(f)
	c.Logger.
		WithField("src", dataUrl).
		WithField("dst", fileName).
		Infof("Downloading players into a file")
	err = c.Downloader.Download(dataUrl, f)
	_ = f.Close()
	if err != nil {
		return errors.Wrap(err, "error while downloading data")
	}
	c.Logger.Infof("Players have been fetched to a file: %s", fileName)
	r, err := c.GzCsvReaderFactory(fileName)
	if err != nil {
		return errors.Wrap(err, "error while creating/initializing gz-csv-reader")
	}
	defer r.Close()
	queue := make(chan Player)
	c.Logger.WithField("total", c.Concurrency).Debugf("Starting workers ...")
	for j := 1; j <= c.Concurrency; j++ {
		c.Logger.WithField("total", c.Concurrency).Debugf("Starting worker #%d ...", j)
		go c.startWorker(j, queue)
	}
	c.Logger.WithField("total", c.Concurrency).Debugf("Workers have been started")
	c.Logger.Infof("Starting players reading ...")
	i := 0
	for {
		i += 1
		c.Logger.Debugf("Reading row #%d", i)
		// last_active:2018-10-26 08:48:42
		// id:059f4d57-xxxx-xxxx-xxxx-fa83792fd276
		pd, err := r.ReadLine()
		if err != nil {
			if err == io.EOF {
				c.Logger.Debugf("EOF")
				break
			}
			return errors.Wrapf(err, "error reading line #%d", i)
		}
		c.Logger.Debugf("Row #%d: %v", i, pd)
		p, err := c.unmarshalPlayerData(pd)
		if err != nil {
			c.Logger.
				WithField("id", pd["id"]).
				WithField("last-active", pd["last_active"]).
				WithError(err).
				Errorf("Error while unmarshalling a player data")
			continue
		}
		c.Logger.Debugf("Sending player handling to queue: %s", pd["id"])
		queue <- p
	}
	close(queue)
	return nil
}

func (c *Cleaner) unmarshalPlayerData(pd PlayerData) (Player, error) {
	p := Player{
		Id: pd["id"],
	}
	lastActive, err := time.Parse("2006-01-02 15:04:05", pd["last_active"])
	if err != nil {
		return Player{}, errors.Wrapf(err, "error while parsing last active: %s", pd["last_active"])
	}
	p.LastActive = lastActive
	return p, nil
}

func (c *Cleaner) startWorker(id int, queue <-chan Player) {
	for p := range queue {
		c.Logger.WithField("worker", id).WithField("player", p.Id).Debugf("Starting a player deletion ...")
		c.handlePlayer(p)
		c.Logger.WithField("worker", id).WithField("player", p.Id).Debugf("Player deletion has been finished")
	}
}

func (c *Cleaner) handlePlayer(p Player) {
	if int(p.LastActive.Unix()) > c.Now()-c.TTL {
		c.Logger.
			WithField("id", p.Id).
			WithField("last-active", p.LastActive.String()).
			Infof("Player is active")
		return
	}
	c.Logger.
		WithField("id", p.Id).
		WithField("last-active", p.LastActive.String()).
		Infof("Player is inactive")
	err := c.OneSignalClient.DeletePlayer(p.Id)
	if err != nil {
		c.Logger.
			WithField("id", p.Id).
			WithField("last-active", p.LastActive.String()).
			WithError(err).
			Errorf("Error while deleting a player")
		return
	}
	c.Logger.
		WithField("id", p.Id).
		WithField("last-active", p.LastActive.String()).
		Infof("User has been deleted successfully")
}

func (c *Cleaner) getDestFileName() string {
	now := time.Unix(int64(c.Now()), 0).Format("20060102150405")
	return fmt.Sprintf("%s/onesignal-players-%s-%s.csv.gz", c.TmpDir, c.OneSignalClient.AppId, now)
}
