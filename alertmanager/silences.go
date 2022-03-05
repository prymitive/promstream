package alertmanager

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type SilenceState string

const (
	SilenceStatePending SilenceState = "pending"
	SilenceStateActive  SilenceState = "active"
	SilenceStateExpired SilenceState = "expired"
)

type SilenceStatus struct {
	State SilenceState `json:"state"`
}

type Matcher struct {
	IsEqual bool   `json:"isEqual"`
	IsRegex bool   `json:"isRegex"`
	Name    string `json:"name"`
	Value   string `json:"value"`
}

type Silence struct {
	ID        string        `json:"id,omitempty"`
	Status    SilenceStatus `json:"status"`
	Matchers  []Matcher     `json:"matchers"`
	CreatedBy string        `json:"createdBy"`
	Comment   string        `json:"comment"`
	StartsAt  time.Time     `json:"startsAt"`
	EndsAt    time.Time     `json:"endsAt"`
	UpdatedAt time.Time     `json:"updatedAt"`
}

func (c *Client) Silences() (silences []Silence, err error) {
	var req *http.Request
	req, err = c.newRequest(http.MethodGet, c.silencesPath, nil)

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

	silences = make([]Silence, 0)

	var silence Silence
	for dec.More() {
		silence.Matchers = make([]Matcher, 0, 1)
		err = dec.Decode(&silence)
		if err != nil {
			return nil, fmt.Errorf("cannot parse JSON content: %w", err)
		}
		silences = append(silences, silence)
	}

	return silences, nil
}
