package main

import (
	"bytes"
	"github.com/mingalevme/gologger"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"testing"
)

const TestOnesignalOrigin = "https://my-onesignal-server.off"

func TestOneSignalClient_GetExportUrl(t *testing.T) {
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
	oneSignalClient := NewOneSignalClient("appId", "restApiKey")
	oneSignalClient.OriginUrl = "https://my-onesignal-server.off"
	oneSignalClient.AppHttpClient = appHttpClient
	oneSignalClient.Logger = gologger.NewNullLogger()
	exportUrl, err := oneSignalClient.GetExportUrl()
	assert.Equal(t, "https://onesignal.com/csv_exports/b2f7f966-d8cc-11e4-bed1-df8f05be55ba/users_184948440ec0e334728e87228011ff41_2015-11-10.csv.gz", exportUrl)
	assert.NoError(t, err)
}

func TestOneSignalClient_DeletePlayer(t *testing.T) {
	playerId := "some-player-id"
	appHttpClient := &TestAppHttpClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			assert.Equal(t, TestOnesignalOrigin+"/api/v1/players/" + playerId + "?app_id=appId", req.URL.String())
			assert.Equal(t, "Basic restApiKey", req.Header.Get("Authorization"))
			assert.Equal(t, "application/json", req.Header.Get("Content-Type"))
			assert.Equal(t, "application/json", req.Header.Get("Accept"))
			return &http.Response{
				StatusCode: 200,
				Body:       ioutil.NopCloser(bytes.NewBufferString("{\"success\": true}")),
				Request:    req,
			}, nil
		},
	}
	oneSignalClient := NewOneSignalClient("appId", "restApiKey")
	oneSignalClient.OriginUrl = "https://my-onesignal-server.off"
	oneSignalClient.AppHttpClient = appHttpClient
	oneSignalClient.Logger = gologger.NewNullLogger()
	err := oneSignalClient.DeletePlayer(playerId)
	assert.NoError(t, err)
}
