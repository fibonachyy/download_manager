package idm

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
)

type section struct {
	Start          int64
	End            int64
	ID             int64
	Name           string
	Current        int64
	ParentDownload *Download
}

func (s *section) completePercentage() int64 {
	percentage := ((s.Current - s.Start) / (s.End - s.Start)) * 100
	return percentage
}
func (s *section) Resume(ctx context.Context) error {

	r, err := s.ParentDownload.getNewRequest("GET")

	if err != nil {
		return err
	}

	r.Header.Set("Range", fmt.Sprintf("bytes=%v-%v", s.Current, s.End))
	rWithContext := r.WithContext(ctx)
	fileName := fmt.Sprintf("%v/%v-%v.tmp", s.ParentDownload.SavePath, s.ParentDownload.FileName, s.ID)
	// Create or open the file for writing

	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)

	if err != nil {
		return fmt.Errorf("error creating tmp file: %v", err)
	}
	defer file.Close() // Close the file when the download is complete or if there's an error
	t, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("error reading tmp file: %v", err)
	}
	fmt.Println(t)
	resp, err := http.DefaultClient.Do(rWithContext)
	if err != nil {
		return err
	}
	if resp.StatusCode > 299 {
		return fmt.Errorf("Can't process section %d, response is %v", s.ID, resp.StatusCode)
	}

	var bytesRead int64
	bufferSize := 4096                    // Adjust the buffer size as needed
	buffer := make([]byte, 0, bufferSize) // Initialize an empty buffer with capacity

	progressReader := &progressReader{
		Reader: io.TeeReader(resp.Body, file),
		OnRead: func(n int, p []byte) {
			bytesRead += int64(n)

			// Append the downloaded data to 'buffer'
			buffer = append(buffer, p[:n]...)

			// Check if the buffer is full
			if len(buffer) >= bufferSize {
				// Write the entire buffer to the file
				_, _ = file.Write(buffer)
				s.Current = bytesRead
				buffer = buffer[:0] // Reset the buffer
			}
		},
	}

	_, err = io.Copy(io.Discard, progressReader) // Copy to discard the data and trigger the progress updates

	// Write any remaining data in the buffer to the file
	if len(buffer) > 0 {
		_, _ = file.Write(buffer)
	}

	return nil
}
