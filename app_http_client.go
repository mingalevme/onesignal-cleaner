package main

import "net/http"

type AppHttpClient interface {
	Do(req *http.Request) (*http.Response, error)
}
