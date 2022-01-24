package main

import (
	"bytes"
	"errors"
	"github.com/mingalevme/gologger"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"testing"
)

const TestOnesignalOrigin = "https://my-onesignal-server.off"

type TestAppHttpClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

func (c *TestAppHttpClient) Do(req *http.Request) (*http.Response, error) {
	if c.DoFunc == nil {
		panic(errors.New("\"DoFunc\" has not been initialized"))
	}
	return c.DoFunc(req)
}

func TestGetExportUrl(t *testing.T) {
	appHttpClient := &TestAppHttpClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			assert.Equal(t, TestOnesignalOrigin+"/api/v1/players/csv_export?app_id=appId", req.URL.String())
			assert.Equal(t, "Basic restApiKey", req.Header.Get("Authorization"))
			assert.Equal(t, "application/json", req.Header.Get("Content-Type"))
			assert.Equal(t, "application/json", req.Header.Get("Accept"))
			return &http.Response{
				StatusCode: 200,
				Body:       ioutil.NopCloser(bytes.NewBufferString("{ \"csv_file_url\": \"https://onesignal.com/csv_exports/b2f7f966-d8cc-11e4-bed1-df8f05be55ba/users_184948440ec0e334728e87228011ff41_2015-11-10.csv.gz\" }")),
				Request:    req,
			}, nil
		},
	}
	oneSignalClient := NewOneSignalClient("https://my-onesignal-server.off", "appId", "restApiKey", appHttpClient, gologger.NewStdoutLogger(gologger.LevelDebug))
	exportUrl, err := oneSignalClient.GetExportUrl()
	assert.Equal(t, "https://onesignal.com/csv_exports/b2f7f966-d8cc-11e4-bed1-df8f05be55ba/users_184948440ec0e334728e87228011ff41_2015-11-10.csv.gz", exportUrl)
	assert.NoError(t, err)
}
