package main

import "time"

type Axios struct {
	BaseURL string
	Headers map[string]string
	Timeout time.Duration
}

func NewAxios(baseURL string, timeout time.Duration, headers map[string]string) *Axios {
	return &Axios{
		BaseURL: baseURL,
		Headers: headers,
		Timeout: timeout,
	}
}

type RequestOptions struct {
	Method  string
	URL     string
	Headers map[string]string
	Body    interface{}
}

type Response struct {
	Status     string
	StatusCode int
	Headers    map[string]string
	Body       []byte
}

func parseStatusLine