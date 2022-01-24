package main

import (
	"encoding/json"
	"fmt"
	"github.com/mingalevme/gologger"
	"github.com/pkg/errors"
	"io"
	"io/ioutil"
	"net/http"
	urllib "net/url"
	"os"
	"strconv"
	"time"
)

const OnesignalOrigin = "https://onesignal.com"

type OneSignalClient struct {
	OriginUrl  string
	AppId      string
	RestApiKey string
	HttpClient AppHttpClient
	Logger     gologger.Logger
	Now        func() int
	Timeout    int
	Pause      int
	TmpDir     string
}

func NewOneSignalClient(appId string, restApiKey string) *OneSignalClient {
	return &OneSignalClient{
		OriginUrl:  OnesignalOrigin,
		AppId:      appId,
		RestApiKey: restApiKey,
		HttpClient: http.DefaultClient,
		Logger:     gologger.NewStdoutLogger(gologger.LevelInfo),
		Timeout:    600,
		Pause:      5,
		Now: func() int {
			return int(time.Now().Unix())
		},
		TmpDir: os.TempDir(),
	}
}

func (c *OneSignalClient) FetchPlayers() (string, error) {
	dataUrl, err := c.GetExportUrl()
	if err != nil {
		return "", errors.Wrap(err, "error while getting an export data url")
	}
	dataFileName := c.getDestinationFileName()
	dataFile, err := os.Create(dataFileName)
	if err != nil {
		return "", errors.Wrap(err, "error while creating a temporary file")
	}
	defer func(dataFile *os.File) {
		_ = dataFile.Close()
	}(dataFile)
	res, err := c.getData(dataUrl)
	if err != nil {
		return "", errors.Wrap(err, "error while downloading a data file")
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(res.Body)
	n, err := io.Copy(dataFile, res.Body)
	contentLength, err := strconv.Atoi(res.Header.Get("Content-Length"))
	if err != nil {
		return "", errors.Wrap(err, "error while getting response content length")
	}
	if int(n) != contentLength {
		return "", errors.Errorf("invalid number of bytes were copied while downloading data file: %d insted of %d", n, contentLength)
	}
	return dataFileName, nil
}

func (c *OneSignalClient) Clean() error {
	dataFileName, err := c.FetchPlayers()
	if err != nil {
		return errors.Wrap(err, "error while fetching players")
	}
	defer func(name string) {
		_ = os.Remove(name)
	}(dataFileName)
	c.Logger.Debugf("Players have been fetched: %s", dataFileName)
	reader, err := NewGzCsvReader(dataFileName)
	if err != nil {
		return errors.Wrap(err, "error while creating/initializing gz-csv-reader")
	}
	defer reader.Close()
	i := 0
	for {
		i += 1
		c.Logger.Debugf("Reading row #%d",i)
		// last_active:2018-10-26 08:48:42
		// id:059f4d57-xxxx-xxxx-xxxx-fa83792fd276
		player, err := reader.ReadLine()
		if err != nil {
			if err == io.EOF {
				c.Logger.Debugf("EOF")
				break
			}
			return errors.Wrapf(err, "error reading line #%d", i)
		}
		c.Logger.Debugf("Row #%d: %v", i, player)
	}
	return nil
}

func (c *OneSignalClient) getData(dataUrl string) (*http.Response, error) {
	req := c.createDataRequest(dataUrl)
	startedAt := c.Now()
	c.Logger.
		WithField("url", dataUrl).
		WithField("now", startedAt).
		Infof("Starting download CSV data")
	for ok := true; ok; ok = c.Now() < (startedAt + c.Timeout) {
		c.Logger.
			WithField("url", dataUrl).
			Infof("Requesting OneSignal for players")
		res, err := c.HttpClient.Do(req)
		if err != nil {
			c.Logger.
				WithField("url", dataUrl).
				WithError(err).
				Errorf("Error while requesting OneSignal for players")
		} else if res.StatusCode == 200 {
			c.Logger.
				WithField("url", dataUrl).
				Infof("Data is ready while requesting OneSignal for players")
			return res, nil
		} else if res.StatusCode == 403 {
			c.Logger.
				WithField("url", dataUrl).
				Infof("Data is not ready while requesting OneSignal for players")
		} else {
			c.Logger.
				WithField("url", dataUrl).
				WithField("response-status-code", res.StatusCode).
				Errorf("Unexpected response while requesting OneSignal for players")
		}
		_ = res.Body.Close()
		c.Logger.
			WithField("url", dataUrl).
			Infof("Sleeping %d second(s) while requesting OneSignal for players", c.Pause)
		time.Sleep(time.Duration(c.Pause) * time.Second)
	}
	c.Logger.
		WithField("url", dataUrl).
		WithField("started-at", startedAt).
		WithField("now", c.Now()).
		Errorf("Timeout while requesting OneSignal for players")
	return nil, errors.Errorf("timeout of %d (seconds) has been exceeded while requesting OneSignal for players", c.Timeout)
}

func (c *OneSignalClient) getDestinationDirName() string {
	return os.TempDir()
}

func (c *OneSignalClient) getDestinationFileName() string {
	return c.getDestinationDirName() + "/onesignal-players-" + c.AppId + ".csv.gz"
}

func (c *OneSignalClient) GetExportUrl() (string, error) {
	req := c.createGetExportRequest()
	res, err := c.HttpClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(res.Body)
	if res.StatusCode != 200 {
		body, _ := ioutil.ReadAll(res.Body)
		panic(errors.New(fmt.Sprintf("error response while fetching export url (200): %s", string(body))))
	}
	// { "csv_file_url": "https://onesignal.com/csv_exports/b2f7f966-d8cc-11e4-bed1-df8f05be55ba/users_184948440ec0e334728e87228011ff41_2015-11-10.csv.gz" }
	var body struct {
		CsvFileUrl string `json:"csv_file_url"`
	}
	err = json.NewDecoder(res.Body).Decode(&body)
	if err != nil {
		return "", err
	}
	if body.CsvFileUrl == "" {
		return "", errors.New("empty csv_file_url")
	}
	return body.CsvFileUrl, nil
}

func (c *OneSignalClient) createGetExportRequest() *http.Request {
	endpointUrl, err := urllib.Parse(c.OriginUrl + "/api/v1/players/csv_export")
	if err != nil {
		panic(err)
	}
	q := endpointUrl.Query()
	q.Set("app_id", c.AppId)
	endpointUrl.RawQuery = q.Encode()
	req, err := http.NewRequest(http.MethodPost, endpointUrl.String(), nil)
	if err != nil {
		panic(err)
	}
	req.Header.Add("Authorization", "Basic "+c.RestApiKey)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	return req
}

func (c *OneSignalClient) createDataRequest(dataUrl string) *http.Request {
	req, err := http.NewRequest(http.MethodGet, dataUrl, nil)
	if err != nil {
		panic(err)
	}
	return req
}
