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

func TestSilences(t *testing.T) {
	fixtures := t.TempDir()
	for _, count := range []int{0, 1, 10, 100, 1000, 5000} {
		writeSilences(t, count, path.Join(fixtures, fmt.Sprintf("%d.json", count)))
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
			c, err := alertmanager.NewClient(svr.URL, alertmanager.WithSilencesPath(f))
			if err != nil {
				t.Error(err)
				t.FailNow()
			}
			silences, err := c.Silences()
			if err != nil {
				t.Error(err)
				t.FailNow()
			}

			d, err := ioutil.ReadFile(path.Join(fixtures, f))
			if err != nil {
				t.Error(err)
				t.FailNow()
			}
			var jsonSilences []alertmanager.Silence
			if err = json.Unmarshal(d, &jsonSilences); err != nil {
				t.Error(err)
				t.FailNow()
			}

			if diff := cmp.Diff(jsonSilences, silences); diff != "" {
				t.Errorf("Silences() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
