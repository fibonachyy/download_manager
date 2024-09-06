package idm

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
)

type section struct {
	Start          int
	End            int
	ID             int
	Name           string
	Current        int
	ParentDownload *Download
}

var (
	bufferSize                  = 64
	permisionNumber fs.FileMode = 0666
)

// func (s *section) completePercentage() int {
// 	percentage := ((s.Current - s.Start) / (s.End - s.Start)) * 100
// 	return percentage
// }

func (s *section) fileName() string {
	return fmt.Sprintf("%v/%v-%v.tmp", s.ParentDownload.SavePath, s.ParentDownload.FileName, s.ID)
}

func (s *section) size() int {
	return s.End - s.Start
}

func (s *section) checkChunkFile() (int, error) {
	// Get file information
	fileInfo, err := os.Stat(s.fileName())
	if err != nil {
		return 0, err
	}
	return int(fileInfo.Size()), nil
}

func (s *section) Resume(ctx context.Context) error {
	fmt.Println("section resumed")
	// Create or open the file for writing
	downloadedBytes, _ := s.checkChunkFile()
	if downloadedBytes == s.size() {
		fmt.Printf("Section %d has already been fully downloaded\n", s.ID)
		return nil
	}
	// If there's partial download, calculate the current range
	s.Current = s.Start + downloadedBytes

	fmt.Printf("Resuming download for section %d: resume from %d\n", s.ID, s.Current)

	fmt.Printf("%#v-%v-%v\n", s, downloadedBytes, s.size())
	r, err := s.ParentDownload.getNewRequest("GET")
	if err != nil {
		return err
	}
	r.Header.Set("Range", fmt.Sprintf("bytes=%v-%v", s.Current, s.End))
	rWithContext := r.WithContext(ctx)

	file, err := os.OpenFile(s.fileName(), os.O_CREATE|os.O_APPEND, permisionNumber)

	if err != nil {
		return fmt.Errorf("error creating tmp file: %v", err)
	}
	defer file.Close() // Close the file when the download is complete or if there's an error
	file.Seek(int64(downloadedBytes), io.SeekStart)
	resp, err := http.DefaultClient.Do(rWithContext)
	if err != nil {
		return err
	}
	if resp.StatusCode > 299 {
		return fmt.Errorf("can't process section %d, response is %v", s.ID, resp.StatusCode)
	}

	var bytesRead int                     // Adjust the buffer size as needed
	buffer := make([]byte, 0, bufferSize) // Initialize an empty buffer with capacity

	progressReader := &progressReader{
		Reader: io.TeeReader(resp.Body, file),
		OnRead: func(n int, p []byte) {
			bytesRead += int(n)

			// Append the downloaded data to 'buffer'
			buffer = append(buffer, p[:n]...)

			// Check if the buffer is full
			if len(buffer) >= bufferSize {
				// Write the entire buffer to the file
				wb, _ := file.Write(buffer)
				fmt.Printf("%v wrote", wb)
				s.Current = bytesRead
				buffer = buffer[:0] // Reset the buffer
			}
		},
	}

	_, err = io.Copy(io.Discard, progressReader) // Copy to discard the data and trigger the progress updates
	if err != nil {
		return fmt.Errorf("discard data failed %v", err)
	}
	// Write any remaining data in the buffer to the file
	if len(buffer) > 0 {
		_, _ = file.Write(buffer)
	}

	return nil
}
