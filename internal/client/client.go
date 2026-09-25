package client

import (
	"errors"
	"net/http"
	"time"
)

var ErrNotHTML = errors.New("not html")

type Client struct {
	httpClient *http.Client
	sem        chan struct{}
	ReqTimeout time.Duration
}

func New(maxConn int, reqTimeout time.Duration) Client {
	return Client{
		httpClient: &http.Client{
			Timeout: reqTimeout,
			Transport: &http.Transport{
				MaxIdleConns:          100,
				MaxIdleConnsPerHost:   50,
				IdleConnTimeout:       90 * time.Second,
				TLSHandshakeTimeout:   10 * time.Second,
				ResponseHeaderTimeout: 15 * time.Second,
			},
		},
		sem: make(chan struct{}, maxConn),
	}
}
