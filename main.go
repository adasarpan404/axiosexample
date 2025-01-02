package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

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

func parseStatusLine(statusLine string) (string, int, error) {
	parts := strings.SplitN(statusLine, " ", 3)
	if len(parts) < 2 {
		return "", 0, errors.New("malformed status line")
	}
	code, err := strconv.Atoi(parts[1])
	if err != nil {
		return "", 0, err
	}
	return parts[0] + " " + parts[1], code, nil
}

func (a *Axios) Request(options RequestOptions) (*Response, error) {
	host := strings.TrimPrefix(a.BaseURL, "http://")
	if strings.Contains(host, "/") {
		host = strings.Split(host, "/")[0]
	}

	url := options.URL
	if !strings.HasPrefix(url, "/") {
		url = "/" + url
	}

	var body []byte
	if options.Body != nil {
		var err error
		body, err = json.Marshal(options.Body)
		if err != nil {
			return nil, errors.New("failed to serialize request body")
		}
	}

	conn, err := net.DialTimeout("tcp", host+":80", a.Timeout)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	reqBuffer := bytes.Buffer{}
	reqBuffer.WriteString(options.Method + " " + url + " HTTP/1.1\r\n")
	reqBuffer.WriteString("Host: " + host + "\r\n")
	for key, value := range a.Headers {
		reqBuffer.WriteString(key + ": " + value + "\r\n")
	}
	for key, value := range options.Headers {
		reqBuffer.WriteString(key + ": " + value + "\r\n")
	}
	if len(body) > 0 {
		reqBuffer.WriteString("Content-Length: " + strconv.Itoa(len(body)) + "\r\n")
	}
	reqBuffer.WriteString("Connection: close\r\n")
	reqBuffer.WriteString("\r\n")
	if len(body) > 0 {
		reqBuffer.Write(body)
	}

	_, err = conn.Write(reqBuffer.Bytes())
	if err != nil {
		return nil, err
	}

	respBuffer := make([]byte, 8192)
	n, err := conn.Read(respBuffer)
	if err != nil {
		return nil, err
	}

	responseString := string(respBuffer[:n])
	parts := strings.SplitN(responseString, "\r\n\r\n", 2)
	if len(parts) < 2 {
		return nil, errors.New("malformed HTTP response")
	}

	headersAndStatus := strings.SplitN(parts[0], "\r\n", 2)
	if len(headersAndStatus) < 2 {
		return nil, errors.New("malformed HTTP response headers")
	}

	statusLine := headersAndStatus[0]
	status, statusCode, err := parseStatusLine(statusLine)
	if err != nil {
		return nil, err
	}

	headers := make(map[string]string)
	headerLines := strings.Split(headersAndStatus[1], "\r\n")
	for _, line := range headerLines {
		parts := strings.SplitN(line, ": ", 2)
		if len(parts) == 2 {
			headers[parts[0]] = parts[1]
		}
	}

	return &Response{
		Status:     status,
		StatusCode: statusCode,
		Headers:    headers,
		Body:       []byte(parts[1]),
	}, nil
}

func main() {
	client := NewAxios("http://jsonplaceholder.typicode.com", 10*time.Second, map[string]string{
		"Content-Type": "application/json",
	})

	response, err := client.Request(RequestOptions{
		Method:  "GET",
		URL:     "/posts/1",
		Headers: nil,
	})
	if err != nil {
		panic(err)
	}
	fmt.Println("GET Response Status:", response.Status)
	fmt.Println("GET Response Body:", string(response.Body))
}
