package main

import (
	"encoding/json"
	"github.com/mingalevme/gologger"
	"github.com/pkg/errors"
	"io"
	"io/ioutil"
	"net/http"
	urllib "net/url"
)

const OnesignalOrigin = "https://onesignal.com"

type OneSignalClient struct {
	OriginUrl     string
	AppId         string
	RestApiKey    string
	AppHttpClient AppHttpClient
	Logger        gologger.Logger
}

func NewOneSignalClient(appId string, restApiKey string) *OneSignalClient {
	return &OneSignalClient{
		OriginUrl:     OnesignalOrigin,
		AppId:         appId,
		RestApiKey:    restApiKey,
		AppHttpClient: http.DefaultClient,
		Logger:        gologger.NewStdoutLogger(gologger.LevelInfo),
	}
}

func (c *OneSignalClient) GetExportUrl() (string, error) {
	req := c.createGetExportRequest()
	resp, err := c.AppHttpClient.Do(req)
	if err != nil {
		return "", errors.Wrapf(err, "error while requesting export url")
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)
	if resp.StatusCode != 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		return "", errors.Errorf("error response while requesting export url (%d): %s", resp.StatusCode, string(body))
	}
	// { "csv_file_url": "https://onesignal.com/csv_exports/b2f7f966-d8cc-11e4-bed1-df8f05be55ba/users_184948440ec0e334728e87228011ff41_2015-11-10.csv.gz" }
	var body struct {
		CsvFileUrl string `json:"csv_file_url"`
	}
	err = json.NewDecoder(resp.Body).Decode(&body)
	if err != nil {
		return "", errors.Wrapf(err, "error while body json-decoding")
	}
	if body.CsvFileUrl == "" {
		return "", errors.New("empty csv_file_url")
	}
	return body.CsvFileUrl, nil
}

func (c *OneSignalClient) DeletePlayer(id string) error {
	req := c.createDeletePlayerRequest(id)
	res, err := c.AppHttpClient.Do(req)
	if err != nil {
		return errors.Wrapf(err, "error while requesting a player deletion: %s", id)
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(res.Body)
	if res.StatusCode != 200 {
		body, _ := ioutil.ReadAll(res.Body)
		return errors.Errorf("error response (code: %d) while requesting a player deletion (%s): %s", res.StatusCode, id, string(body))
	}
	// {'success': true}
	var body struct {
		Success bool `json:"success"`
	}
	err = json.NewDecoder(res.Body).Decode(&body)
	if err != nil {
		return errors.Wrapf(err, "error while decoding a json-body while requesting a player deletion: %s", id)
	}
	if !body.Success {
		return errors.Errorf("error response payload while requesting a player deletion (%s): %v", id, body)
	}
	return nil
}

func (c *OneSignalClient) createGetExportRequest() *http.Request {
	return c.createRequest(http.MethodPost, "/api/v1/players/csv_export")
}

func (c *OneSignalClient) createDeletePlayerRequest(id string) *http.Request {
	return c.createRequest(http.MethodDelete, "/api/v1/players/"+id)
}

func (c *OneSignalClient) createRequest(method string, path string) *http.Request {
	endpointUrl, err := urllib.Parse(c.OriginUrl + path)
	if err != nil {
		panic(err)
	}
	q := endpointUrl.Query()
	q.Set("app_id", c.AppId)
	endpointUrl.RawQuery = q.Encode()
	req, err := http.NewRequest(method, endpointUrl.String(), nil)
	if err != nil {
		panic(err)
	}
	req.Header.Add("Authorization", "Basic "+c.RestApiKey)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	return req
}
