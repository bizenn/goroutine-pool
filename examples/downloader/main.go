package main

import (
	"bufio"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	gpool "github.com/bizenn/goroutine-pool"
)

// DownloadJob ...
type DownloadJob struct {
	gpool.Job
	client *http.Client
	url    string
}

// NewDownloadJob ...
func NewDownloadJob(c *http.Client, u string) *DownloadJob {
	var j = &DownloadJob{
		client: c,
		url:    u,
	}
	return j
}

// Do ...
func (j *DownloadJob) Do() {
	var u *url.URL
	var err error
	if u, err = url.Parse(j.url); err == nil {
		var dest = filepath.Join(".", u.Path)
		var dir = filepath.Dir(dest)
		if err = os.MkdirAll(dir, 0777); err == nil {
			var res *http.Response
			if res, err = j.client.Get(j.url); err == nil {
				if res.StatusCode != http.StatusOK {
					log.Printf("%s: %q", res.Status, j.url)
				} else {
					var out *os.File
					if out, err = ioutil.TempFile(dir, filepath.Base(dest)); err == nil {
						if _, err = io.Copy(out, res.Body); err == nil {
							err = os.Rename(out.Name(), dest)
						}
						_ = out.Close()
						_ = os.Remove(out.Name())
					}
				}
				_, _ = io.Copy(ioutil.Discard, res.Body)
				_ = res.Body.Close()
			}
		}
	}
	if err != nil {
		log.Printf("%s: %q", err, j.url)
	} else {
		log.Printf("Downloaded: %q", j.url)
	}
}

func main() {
	var gp = gpool.NewPool(2) // 2 goroutine started.
	var in = bufio.NewScanner(os.Stdin)
	for in.Scan() {
		gp.Channel() <- NewDownloadJob(http.DefaultClient, in.Text())
	}
	gp.Shutdown()
	gp.Wait()
}
