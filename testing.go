// +build testing

package main

import (
	"github.com/pkg/errors"
	"net/http"
)

type TestAppHttpClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

func (c *TestAppHttpClient) Do(req *http.Request) (*http.Response, error) {
	if c.DoFunc == nil {
		panic(errors.New("\"DoFunc\" has not been initialized"))
	}
	return c.DoFunc(req)
}

type QueueResponseAppHttpClient struct {
	Responses []*http.Response
}

func NewQueueResponseAppHttpClient() *QueueResponseAppHttpClient {
	return &QueueResponseAppHttpClient{
		Responses: []*http.Response{},
	}
}

func (c *QueueResponseAppHttpClient) Enqueue(resp *http.Response) {
	c.Responses = append(c.Responses, resp)
}

func (c *QueueResponseAppHttpClient) Dequeue() *http.Response {
	if c.Size() == 0 {
		panic(errors.Errorf("queue is empty"))
	}
	resp := c.Responses[0]
	c.Responses = c.Responses[1:]
	return resp
}

func (c *QueueResponseAppHttpClient) Size() int {
	return len(c.Responses)
}

func (c *QueueResponseAppHttpClient) Do(req *http.Request) (*http.Response, error) {
	if c.Size() == 0 {
		panic(errors.Errorf("unexpected request: %v", req))
	}
	return c.Dequeue(), nil
}