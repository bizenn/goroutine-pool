# gpool: goroutine pool library

The goroutine-pool(gpool) aims to provide goroutine pool library,
arbitrary fixed number of goroutine to process "job".

## Installation

``` shellsession
% go get -u github.com/bizenn/goroutine-pool
```

## Required

- logrus: github.com/sirupsen/logrus

## Full Example

A downloader. This program read each line as a URL from standard input,
and GET and save it to local storage concurrently.  The concurrency is
just 2.

``` go
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
	gpool.BaseJob
	*sync.WaitGroup
	client *http.Client
	url    string
}

// NewDownloadJob ...
func NewDownloadJob(wg *sync.WaitGroup, c *http.Client, u string) *DownloadJob {
	var j = &DownloadJob{
		WaitGroup: wg,
		client:    c,
		url:       u,
	}
	j.Add(1)
	return j
}

// LogProperties ...
func (j *DownloadJob) LogProperties() logrus.Fields {
	return logrus.Fields{
		"url": j.url,
	}
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
			} else {
				io.Copy(ioutil.Discard, res.Body)
			}
			res.Body.Close()
		}
	}
	if err != nil {
		logrus.WithFields(j.LogProperties()).WithError(err).Warn("failed")
	} else {
		logrus.WithFields(j.LogProperties()).Info("done")
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
```
