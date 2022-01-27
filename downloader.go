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
	AppHttpClient    AppHttpClient
	ReadinessTimeout int
	Pause            time.Duration
	Now               Nower
	Logger            gologger.Logger
}

func NewDownloader() *Downloader {
	return &Downloader{
		AppHttpClient:    http.DefaultClient,
		ReadinessTimeout: 600,
		Pause:            5 * time.Second,
		Now:              Now,
		Logger:           gologger.NewStdoutLogger(gologger.LevelInfo),
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
	contentLength, err := strconv.Atoi(resp.Header.Get("Content-Length"))
	if err != nil {
		return errors.Wrap(err, "error while getting response Content-Length-header")
	}
	d.Logger.
		WithField("url", sourceURL).
		WithField("total", contentLength).
		Infof("Reading response body into a destination ...")
	wc := &WriteCounter{
		Total:  contentLength,
		Logger: d.Logger,
	}
	tee := io.TeeReader(resp.Body, wc)
	n, err := io.Copy(destination, tee)
	if int(n) != contentLength {
		d.Logger.
			WithField("url", sourceURL).
			WithField("io.Copy", n).
			WithField("WriteCounter", wc.Received).
			WithField("Content-Length", resp.Header.Get("Content-Length")).
			Warningf("Invalid number of bytes were written while downloading a remote data")
		return nil
	}
	d.Logger.WithField("url", sourceURL).Infof("Remote resource has been successfully downloaded")
	return nil
}

func (d *Downloader) request(sourceURL string) (*http.Response, error) {
	req := d.createRequest(sourceURL)
	startedAt := d.Now()
	attempt := 0
	d.Logger.
		WithField("url", sourceURL).
		WithField("now", startedAt).
		Infof("Requesting a remote data ...")
	for ok := true; ok; ok = d.Now() < (startedAt + d.ReadinessTimeout) {
		attempt += 1
		d.Logger.
			WithField("url", sourceURL).
			WithField("attempt", attempt).
			Debugf("Requesting a remote data")
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
				Debugf("Data is not ready while requesting a remote data")
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
			Debugf("Sleeping %s while requesting a remote data", d.Pause)
		time.Sleep(d.Pause)
	}
	d.Logger.
		WithField("url", sourceURL).
		WithField("started-at", startedAt).
		WithField("now", d.Now()).
		Errorf("ReadinessTimeout while requesting a remote data")
	return nil, errors.Errorf("ReadinessTimeout of %d (seconds) has been exceeded while requesting a remote data", d.ReadinessTimeout)
}

func (d *Downloader) createRequest(sourceURL string) *http.Request {
	req, err := http.NewRequest(http.MethodGet, sourceURL, nil)
	if err != nil {
		panic(err)
	}
	return req
}

type WriteCounter struct {
	Total    int
	Received int
	Logger   gologger.Logger
}

func (wc *WriteCounter) Write(p []byte) (int, error) {
	n := len(p)
	wc.Received += n
	wc.Logger.
		WithField("size", n).
		WithField("received", wc.Received).
		WithField("total", wc.Total).
		Debugf("Data chunk received")
	return n, nil
}
