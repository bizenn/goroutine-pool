package main

import (
	"bufio"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sync"

	"github.com/bizenn/goroutine-pool"
	"github.com/sirupsen/logrus"
)

// DownloadJob ...
type DownloadJob struct {
	gpool.Job
	*sync.WaitGroup
	client  *http.Client
	url     string
	logging *logrus.Entry
}

// NewDownloadJob ...
func NewDownloadJob(wg *sync.WaitGroup, c *http.Client, u string) *DownloadJob {
	var j = &DownloadJob{
		WaitGroup: wg,
		client:    c,
		url:       u,
		logging:   logrus.WithField("url", u),
	}
	j.Add(1)
	return j
}

// Do ...
func (j *DownloadJob) Do() (err error) {
	defer j.Done()
	var u *url.URL
	if u, err = url.Parse(j.url); err == nil {
		var dest = filepath.Join(".", u.Path)
		var dir = filepath.Dir(dest)
		os.MkdirAll(dir, 0777)
		var res *http.Response
		if res, err = j.client.Get(j.url); err == nil {
			if res.StatusCode == http.StatusOK {
				var out *os.File
				if out, err = ioutil.TempFile(dir, filepath.Base(dest)); err == nil {
					if _, err = io.Copy(out, res.Body); err == nil {
						err = os.Rename(out.Name(), dest)
					}
					out.Close()
					os.Remove(out.Name())
				}
			}
			io.Copy(ioutil.Discard, res.Body)
			res.Body.Close()
		}
	}
	if err != nil {
		j.logging.WithError(err).Warn("failed")
	} else {
		j.logging.Info("done")
	}
	return err
}

func main() {
	var wg sync.WaitGroup
	var gp = gpool.NewPool(2) // 2 goroutine started.
	var in = bufio.NewScanner(os.Stdin)
	for in.Scan() {
		gp.Channel() <- NewDownloadJob(&wg, http.DefaultClient, in.Text())
	}
	wg.Wait()
	gp.Shutdown()
}
