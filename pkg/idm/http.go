package idm

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
)

// Download a single section and save content to a tmp file
func (d *Download) downloadSection(ctx context.Context, section [3]int) error {
	r, err := d.getNewRequest("GET")

	if err != nil {
		return err
	}
	r.Header.Set("Range", fmt.Sprintf("bytes=%v-%v", section[1], section[2]))
	rWithContext := r.WithContext(ctx)
	resp, err := http.DefaultClient.Do(rWithContext)
	if err != nil {
		return err
	}
	if resp.StatusCode > 299 {
		return errors.New(fmt.Sprintf("Can't process, response is %v", resp.StatusCode))
	}
	fmt.Printf("Downloaded %v bytes for section %v\n", resp.Header.Get("Content-Length"), section[0])
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	err = os.WriteFile(fmt.Sprintf("section-%v.tmp", section[0]), b, os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}

// // Get a new http request
func (d *Download) getNewRequest(method string) (*http.Request, error) {
	r, err := http.NewRequest(
		method,
		d.Url,
		nil,
	)
	if err != nil {
		return nil, err
	}
	r.Header.Set("User-Agent", "Silly Download Manager v001")
	return r, nil
}
