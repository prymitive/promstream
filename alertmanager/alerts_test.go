package alertmanager_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/NYTimes/gziphandler"
	"github.com/google/go-cmp/cmp"
	"github.com/prymitive/promstream/alertmanager"
)

func TestAlerts(t *testing.T) {
	fixtures := t.TempDir()
	for _, count := range []int{0, 1, 10, 100, 1000, 5000} {
		writeAlerts(t, count, path.Join(fixtures, fmt.Sprintf("%d.json", count)))
	}

	fs := http.FileServer(http.Dir(fixtures))
	gz, err := gziphandler.GzipHandlerWithOpts(
		gziphandler.ContentTypes([]string{"application/json"}),
		gziphandler.MinSize(1),
	)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	svr := httptest.NewServer(gz(fs))
	defer svr.Close()

	var files []string
	err = filepath.Walk(fixtures, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			files = append(files, filepath.Base(path))
		}
		return nil
	})
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	for _, f := range files {
		t.Run(f, func(t *testing.T) {
			c, err := alertmanager.NewClient(svr.URL, alertmanager.WithAlertsPath(f))
			if err != nil {
				t.Error(err)
				t.FailNow()
			}
			alerts, err := c.Alerts()
			if err != nil {
				t.Error(err)
				t.FailNow()
			}

			d, err := ioutil.ReadFile(path.Join(fixtures, f))
			if err != nil {
				t.Error(err)
				t.FailNow()
			}
			var jsonAlerts []alertmanager.Alert
			if err = json.Unmarshal(d, &jsonAlerts); err != nil {
				t.Error(err)
				t.FailNow()
			}

			if diff := cmp.Diff(jsonAlerts, alerts); diff != "" {
				t.Errorf("Alerts() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
