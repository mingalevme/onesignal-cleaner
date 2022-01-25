package main

import (
	"github.com/mingalevme/gologger"
	"github.com/pkg/errors"
	"io"
	"net/http"
	"strconv"
	"time"
)

type Downloader struct {
	AppHttpClient AppHttpClient
	Timeout       int
	Pause  time.Duration
	Now    Nower
	Logger gologger.Logger
}

func NewDownloader() *Downloader {
	return &Downloader{
		AppHttpClient: http.DefaultClient,
		Timeout:       600,
		Pause:         5 * time.Second,
		Now:           Now,
		Logger:        gologger.NewStdoutLogger(gologger.LevelInfo),
	}
}

func (d *Downloader) Download(sourceURL string, destination io.Writer) error {
	resp, err := d.request(sourceURL)
	if err != nil {
		return errors.Wrap(err, "error while downloading a data file")
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)
	n, err := io.Copy(destination, resp.Body)
	contentLength, err := strconv.Atoi(resp.Header.Get("Content-Length"))
	if err != nil {
		return errors.Wrap(err, "error while getting response Content-Length-header")
	}
	if int(n) != contentLength {
		return errors.Errorf("invalid number of bytes were copied while downloading a remote data: %d insted of %d", n, contentLength)
	}
	return nil
}

func (d *Downloader) request(sourceURL string) (*http.Response, error) {
	req := d.createRequest(sourceURL)
	startedAt := d.Now()
	attempt := 1
	d.Logger.
		WithField("url", sourceURL).
		WithField("now", startedAt).
		Infof("Starting download a remote data")
	for ok := true; ok; ok = d.Now() < (startedAt + d.Timeout) {
		d.Logger.
			WithField("url", sourceURL).
			WithField("attempt", attempt).
			Infof("Requesting a remote data")
		res, err := d.AppHttpClient.Do(req)
		if err != nil {
			d.Logger.
				WithField("url", sourceURL).
				WithField("attempt", attempt).
				WithError(err).
				Errorf("Error while requesting a remote data")
		} else if res.StatusCode == 200 {
			d.Logger.
				WithField("url", sourceURL).
				WithField("attempt", attempt).
				Infof("Data is ready while requesting a remote data")
			return res, nil
		} else if res.StatusCode == 403 {
			d.Logger.
				WithField("url", sourceURL).
				WithField("attempt", attempt).
				Infof("Data is not ready while requesting a remote data")
		} else {
			d.Logger.
				WithField("url", sourceURL).
				WithField("attempt", attempt).
				WithField("response-status-code", res.StatusCode).
				Errorf("Unexpected response while requesting a remote data")
		}
		_ = res.Body.Close()
		d.Logger.
			WithField("url", sourceURL).
			WithField("attempt", attempt).
			Infof("Sleeping %s while requesting a remote data", d.Pause)
		time.Sleep(d.Pause)
	}
	d.Logger.
		WithField("url", sourceURL).
		WithField("started-at", startedAt).
		WithField("now", d.Now()).
		Errorf("Timeout while requesting a remote data")
	return nil, errors.Errorf("timeout of %d (seconds) has been exceeded while requesting a remote data", d.Timeout)
}

func (d *Downloader) createRequest(sourceURL string) *http.Request {
	req, err := http.NewRequest(http.MethodGet, sourceURL, nil)
	if err != nil {
		panic(err)
	}
	return req
}
