package main

import (
	"bytes"
	"github.com/mingalevme/gologger"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
)

func TestCleaner_Clean(t *testing.T) {
	logger := gologger.NewNullLogger()

	// OneSignal
	oneSignalAppHttpClient := NewQueueResponseAppHttpClient()
	oneSignalAppHttpClient.Enqueue(&http.Response{
		StatusCode: 200,
		Body:       ioutil.NopCloser(bytes.NewBufferString("{ \"csv_file_url\": \"https://onesignal.com/csv_exports/b2f7f966-d8cc-11e4-bed1-df8f05be55ba/users_184948440ec0e334728e87228011ff41_2015-11-10.csv.gz\" }")),
	})
	oneSignalAppHttpClient.Enqueue(&http.Response{
		StatusCode: 200,
		Body:       ioutil.NopCloser(bytes.NewBufferString("{\"success\":true}")),
	})
	oneSignal := NewOneSignalClient("app-id", "rest-api-key")
	oneSignal.AppHttpClient = oneSignalAppHttpClient
	oneSignal.Logger = logger

	// Downloader
	downloaderAppHttpClient := NewQueueResponseAppHttpClient()
	downloaderAppHttpClient.Enqueue(&http.Response{
		StatusCode: 200,
		Body:       ioutil.NopCloser(bytes.NewBufferString("foobar")),
		Header: map[string][]string{
			"Content-Length": {"6"},
		},
	})
	downloader := NewDownloader()
	downloader.AppHttpClient = downloaderAppHttpClient
	downloader.Logger = logger

	gzCsvReaderFactory := func(filename string) (*GzCsvReader, error) {
		dir, _ := os.Getwd()
		dataFileName := dir + "/gz_csv_reader_test_data.csv.gz"
		return NewGzCsvReader(dataFileName)
	}

	cleaner := NewCleaner("app-id", "rest-api-key", logger)
	cleaner.OneSignalClient = oneSignal
	cleaner.Downloader = downloader
	cleaner.GzCsvReaderFactory = gzCsvReaderFactory

	err := cleaner.Clean()
	assert.NoError(t, err)

	assert.Equal(t, 0, oneSignalAppHttpClient.Size())
}


