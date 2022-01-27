package main

import (
	"fmt"
	"github.com/mingalevme/gologger"
	"github.com/pkg/errors"
	"io"
	"os"
	"sync"
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
		TTL:         86400 * 30 * 6,
		TmpDir:      os.TempDir(),
		Concurrency: 1,
		Now:         Now,
		Logger:      logger,
	}
}

func (c *Cleaner) Clean(localFileName ...string) error {
	var fileName string
	var err error
	if len(localFileName) > 0 && localFileName[0] != "" {
		fileName = localFileName[0]
		c.Logger.WithField("file", fileName).Infof("Reading data from a local file")
	} else {
		fileName, err = c.fetchData()
		if err != nil {
			return errors.Wrap(err, "error while fetching a data file")
		}
	}
	c.Logger.Infof("Starting data file reading ...")
	r, err := c.GzCsvReaderFactory(fileName)
	if err != nil {
		return errors.Wrap(err, "error while creating/initializing gz-csv-reader")
	}
	defer r.Close()
	throttle := make(chan struct{}, c.Concurrency)
	wg := sync.WaitGroup{}
	c.Logger.Infof("Starting players handling ...")
	i := 0
	for {
		i += 1
		c.Logger.Debugf("Reading row #%d", i)
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
		if int(p.LastActive.Unix()) > c.Now()-c.TTL {
			c.Logger.
				WithField("id", p.Id).
				WithField("last-active", p.LastActive.String()).
				Infof("Player is active")
			continue
		}
		c.Logger.
			WithField("id", p.Id).
			WithField("last-active", p.LastActive.String()).
			Infof("Player is inactive")
		c.Logger.Debugf("Scheduling player for a deletion: %s", pd["id"])
		throttle <- struct{}{}
		wg.Add(1)
		go func(n int) {
			c.Logger.WithField("player", p.Id).Debugf("Starting a player deletion ...")
			c.deletePlayer(p)
			c.Logger.WithField("player", p.Id).Debugf("Player deletion has been finished")
			<-throttle
			wg.Done()
		}(i)
	}
	wg.Wait()
	close(throttle)
	return nil
}

func (c *Cleaner) fetchData() (string, error) {
	dataUrl, err := c.OneSignalClient.GetExportUrl()
	if err != nil {
		return "", errors.Wrap(err, "error while getting export url")
	}
	c.Logger.Infof("Export url has been fetched: %s", dataUrl)
	fileName := c.getDestFileName()
	f, err := os.Create(fileName)
	if err != nil {
		return "", errors.Wrap(err, "error while creating a temporary file")
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
		return "", errors.Wrap(err, "error while downloading data")
	}
	c.Logger.Infof("Players have been fetched to a file: %s", fileName)
	return fileName, nil
}

func (c *Cleaner) unmarshalPlayerData(pd PlayerData) (Player, error) {
	// last_active:2018-10-26 08:48:42
	// id:059f4d57-xxxx-xxxx-xxxx-fa83792fd276
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

func (c *Cleaner) deletePlayer(p Player) {
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
