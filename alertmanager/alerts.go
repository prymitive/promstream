package alertmanager

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type AlertState string

const (
	AlertStateUnprocessed AlertState = "unprocessed"
	AlertStateActive      AlertState = "active"
	AlertStateExpired     AlertState = "expired"
)

type Receiver struct {
	Name string `json:"name"`
}

type AlertStatus struct {
	InhibitedBy []string   `json:"inhibitedBy"`
	SilencedBy  []string   `json:"silencedBy"`
	State       AlertState `json:"state"`
}

type Alert struct {
	Fingerprint  string            `json:"fingerprint"`
	Receivers    []Receiver        `json:"receivers"`
	Status       AlertStatus       `json:"status"`
	StartsAt     time.Time         `json:"startsAt"`
	EndsAt       time.Time         `json:"endsAt"`
	UpdatedAt    time.Time         `json:"updatedAt"`
	GeneratorURL string            `json:"generatorURL"`
	Annotations  map[string]string `json:"annotations"`
	Labels       map[string]string `json:"labels"`
}

func (c *Client) Alerts() (alerts []Alert, err error) {
	var req *http.Request
	req, err = c.newRequest(http.MethodGet, c.alertsPath, nil)

	var resp *http.Response
	if resp, err = c.do(req); err != nil {
		return nil, err
	}
	defer discardBody(resp)

	var r io.Reader
	if r, err = readBody(resp); err != nil {
		return nil, err
	}

	dec := json.NewDecoder(r)

	var tok json.Token
	if tok, err = dec.Token(); err != nil {
		return nil, err
	}
	if tok != json.Delim('[') {
		return nil, fmt.Errorf("cannot parse JSON tokens, expected [, got %s", tok)
	}

	alerts = make([]Alert, 0)

	var alert Alert
	for dec.More() {
		alert.Status = AlertStatus{
			InhibitedBy: make([]string, 0),
			SilencedBy:  make([]string, 0),
			State:       "",
		}
		alert.Receivers = make([]Receiver, 0)
		alert.Annotations = map[string]string{}
		alert.Labels = map[string]string{}
		err = dec.Decode(&alert)
		if err != nil {
			return nil, fmt.Errorf("cannot parse JSON content: %w", err)
		}
		alerts = append(alerts, alert)
	}
	return alerts, nil
}
