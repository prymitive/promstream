package alertmanager_test

import (
	"encoding/json"
	"io/ioutil"
	"math/rand"
	"testing"
	"time"

	"github.com/prymitive/promstream/alertmanager"
)

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func randStrings(n, maxLenght int) []string {
	s := make([]string, 0, n)
	for i := 0; i < n; i++ {
		l := rand.Intn(maxLenght)
		s = append(s, randString(l))
	}
	return s
}

func maybe(n int) bool {
	return rand.Intn(100) <= n
}

func randomDate() time.Time {
	min := time.Date(1970, 1, 0, 0, 0, 0, 0, time.UTC).Unix()
	max := time.Date(2070, 1, 0, 0, 0, 0, 0, time.UTC).Unix()
	delta := max - min

	sec := rand.Int63n(delta) + min
	return time.Unix(sec, 0)
}

func mockMatchers(count int) []alertmanager.Matcher {
	matchers := []alertmanager.Matcher{}

	for i := 0; i < count; i++ {
		name := rand.Intn(20)
		value := rand.Intn(40)
		m := alertmanager.Matcher{
			IsEqual: maybe(70),
			IsRegex: maybe(40),
			Name:    randString(name),
			Value:   randString(value),
		}
		matchers = append(matchers, m)
	}

	return matchers
}

func mockSilences(count int) (silences []alertmanager.Silence) {
	silences = []alertmanager.Silence{}
	for i := 0; i < count; i++ {
		state := alertmanager.SilenceStateActive
		if maybe(20) {
			state = alertmanager.SilenceStatePending
		} else if maybe(30) {
			state = alertmanager.SilenceStateExpired
		}

		startsAt := randomDate()
		endsAt := startsAt.Add(time.Minute * time.Duration(rand.Int63n(30000)))

		author := rand.Intn(50)
		words := rand.Intn(1000)

		matchers := rand.Intn(10)

		silence := alertmanager.Silence{
			ID:        randString(32),
			Status:    alertmanager.SilenceStatus{State: state},
			StartsAt:  startsAt,
			EndsAt:    endsAt,
			UpdatedAt: startsAt,
			CreatedBy: randString(author),
			Comment:   randString(words),
			Matchers:  mockMatchers(matchers),
		}
		silences = append(silences, silence)
	}
	return
}

func writeSilences(t *testing.T, count int, path string) {
	s := mockSilences(count)
	d, _ := json.Marshal(s)
	if err := ioutil.WriteFile(path, d, 0644); err != nil {
		t.Error(err)
		t.FailNow()
	}
}

func mockLabels(count int) map[string]string {
	labels := map[string]string{}
	for i := 0; i < count; i++ {
		name := rand.Intn(20)
		value := rand.Intn(60)
		labels[randString(name)] = randString(value)
	}
	return labels
}

func mockAlerts(count int) (alerts []alertmanager.Alert) {
	alerts = []alertmanager.Alert{}
	for i := 0; i < count; i++ {
		state := alertmanager.AlertStateActive
		if maybe(20) {
			state = alertmanager.AlertStateUnprocessed
		} else if maybe(30) {
			state = alertmanager.AlertStateExpired
		}

		startsAt := randomDate()
		endsAt := startsAt.Add(time.Minute * time.Duration(rand.Int63n(30000)))

		generator := rand.Intn(200)
		annotations := rand.Intn(5)
		labels := rand.Intn(15)

		alert := alertmanager.Alert{
			Fingerprint: randString(32),
			Receivers:   nil,
			Status: alertmanager.AlertStatus{
				InhibitedBy: randStrings(rand.Intn(10), 40),
				SilencedBy:  randStrings(rand.Intn(10), 40),
				State:       state,
			},
			StartsAt:     startsAt,
			EndsAt:       endsAt,
			UpdatedAt:    startsAt,
			GeneratorURL: randString(generator),
			Annotations:  mockLabels(annotations),
			Labels:       mockLabels(labels),
		}
		alerts = append(alerts, alert)
	}
	return
}

func writeAlerts(t *testing.T, count int, path string) {
	s := mockAlerts(count)
	d, _ := json.Marshal(s)
	if err := ioutil.WriteFile(path, d, 0644); err != nil {
		t.Error(err)
		t.FailNow()
	}
}
