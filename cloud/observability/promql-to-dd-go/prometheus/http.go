package prometheus

import (
	"bytes"
	"context"
	"net/http"
	"net/url"
	"path"
	"strings"
)

type HttpClient struct {
	Endpoint *url.URL
	Client   *http.Client
}

func NewHttpClient(addr string, httpClient *http.Client) (*HttpClient, error) {
	u, err := url.Parse(addr)
	if err != nil {
		return nil, err
	}
	u.Path = strings.TrimRight(u.Path, "/")

	return &HttpClient{
		Endpoint: u,
		Client:   httpClient,
	}, nil
}

func (c *HttpClient) URL(ep string, args map[string]string) *url.URL {
	p := path.Join(c.Endpoint.Path, ep)

	for arg, val := range args {
		arg = ":" + arg
		p = strings.ReplaceAll(p, arg, val)
	}

	u := *c.Endpoint
	u.Path = p

	return &u
}

func (c *HttpClient) Do(ctx context.Context, req *http.Request) (*http.Response, []byte, error) {
	if ctx != nil {
		req = req.WithContext(ctx)
	}
	resp, err := c.Client.Do(req)
	defer func() {
		if resp != nil {
			resp.Body.Close()
		}
	}()

	if err != nil {
		return nil, nil, err
	}

	var body []byte
	done := make(chan struct{})
	go func() {
		var buf bytes.Buffer
		_, err = buf.ReadFrom(resp.Body)
		body = buf.Bytes()
		close(done)
	}()

	select {
	case <-ctx.Done():
		<-done
		err = resp.Body.Close()
		if err == nil {
			err = ctx.Err()
		}
	case <-done:
	}

	return resp, body, err
}
