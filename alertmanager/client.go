package alertmanager

import (
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

type Option func(c *Client) error

func NewClient(alertmanagerURI string, opts ...Option) (c *Client, err error) {
	c = &Client{
		url:                 alertmanagerURI,
		httpClient:          &http.Client{},
		responseCompression: true,
		silencesPath:        "/api/v2/silences",
		alertsPath:          "/api/v2/alerts",
	}

	for _, opt := range opts {
		err := opt(c)
		if err != nil {
			return nil, err
		}
	}

	return c, err
}

func WithResponseCompression(compress bool) Option {
	return func(c *Client) error {
		c.responseCompression = compress
		return nil
	}
}

func WithRequestTimeout(timeout time.Duration) Option {
	return func(c *Client) error {
		c.httpClient.Timeout = timeout
		return nil
	}
}

func WithSilencesPath(p string) Option {
	return func(c *Client) error {
		c.silencesPath = p
		return nil
	}
}

func WithAlertsPath(p string) Option {
	return func(c *Client) error {
		c.alertsPath = p
		return nil
	}
}

type Client struct {
	url                 string
	httpClient          *http.Client
	responseCompression bool
	silencesPath        string
	alertsPath          string
}

func (c Client) newRequest(method, path string, body io.Reader) (req *http.Request, err error) {
	url := strings.TrimSuffix(c.url, "/") + "/" + strings.TrimPrefix(path, "/")
	req, err = http.NewRequest("GET", url, body)
	if err != nil {
		return
	}

	if c.responseCompression {
		req.Header.Set("Accept-Encoding", "gzip")
	} else {
		req.Header.Set("Accept-Encoding", "identity")
	}

	return
}

func (c Client) do(req *http.Request) (resp *http.Response, err error) {
	resp, err = c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("invalid response for %s, status: %s", req.URL, resp.Status)
	}

	return resp, nil
}

func readBody(resp *http.Response) (r io.Reader, err error) {
	switch resp.Header.Get("Content-Encoding") {
	case "gzip":
		r, err = gzip.NewReader(resp.Body)
		if err != nil {
			return nil, err
		}
	default:
		r = resp.Body
	}
	return r, nil
}

func discardBody(resp *http.Response) {
	io.Copy(ioutil.Discard, resp.Body)
	resp.Body.Close()
}
