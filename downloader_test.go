package main

import (
	"bytes"
	"github.com/mingalevme/gologger"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestDownloader_Download(t *testing.T) {
	var responses []*http.Response
	responses = append(responses, &http.Response{
		StatusCode: 200,
		Body:       ioutil.NopCloser(bytes.NewBufferString("foobar")),
		Header: map[string][]string{
			"Content-Length": {"6"},
		},
	})
	responses = append(responses, &http.Response{
		StatusCode: 403,
		Body:       ioutil.NopCloser(bytes.NewBufferString("403")),
		Header: map[string][]string{
			"Content-Length": {"3"},
		},
	})
	d := NewDownloader()
	d.AppHttpClient = &TestAppHttpClient{DoFunc: func(req *http.Request) (*http.Response, error) {
		if len(responses) == 0 {
			panic(errors.Errorf("unexpected request: %v", req))
		}
		resp := responses[len(responses)-1]
		resp.Request = req
		responses = responses[:len(responses)-1]
		return resp, nil
	}}
	d.Logger = gologger.NewStdoutLogger(gologger.LevelDebug)
	d.Pause = time.Nanosecond
	src := "https://example.com/test"
	dst := &strings.Builder{}
	err := d.Download(src, dst)
	assert.NoError(t, err)
	assert.Len(t, responses, 0)
}
