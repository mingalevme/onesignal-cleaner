package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/mingalevme/gologger"
	"io"
	"io/ioutil"
	"net/http"
	urllib "net/url"
)

const OnesignalOrigin = "https://onesignal.com"

type OneSignalClient struct {
	OriginUrl  string
	AppId      string
	RestApiKey string
	Client     AppHttpClient
	Logger     gologger.Logger
}

func NewOneSignalClient(originUrl string, appId string, restApiKey string, appHttpClient AppHttpClient, logger gologger.Logger) *OneSignalClient {
	if appHttpClient == nil {
		appHttpClient = http.DefaultClient
	}

	if logger == nil {
		logger = gologger.NewNullLogger()
	}

	return &OneSignalClient{
		OriginUrl:  originUrl,
		AppId:      appId,
		RestApiKey: restApiKey,
		Client:     appHttpClient,
		Logger:     logger,
	}
}

func (c *OneSignalClient) GetExportUrl() (string, error) {
	req := c.createGetExportRequest()
	res, err := c.Client.Do(req)
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
