package idm

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type Download struct {
	Url              string
	Size             int
	SavePath         string
	EndDate          time.Time
	Duration         time.Duration
	Sections         [][3]int
	FirstSectionSize int
	Workers          int
	PortionSize      int
}

func NewDownload(url string, firstSectionSize int, portionSize int, workers int) (*Download, error) {
	return &Download{
		Url:              url,
		FirstSectionSize: firstSectionSize,
		PortionSize:      portionSize,
		Workers:          workers,
		Sections:         [][3]int{},
	}, nil
}

func (d *Download) Start() error {
	// Create a context with cancellation for the download process
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // Ensure cancellation when the main function exits

	err := d.CalculateSize()
	if err != nil {
		return fmt.Errorf("Failed to calculate size: %v", err)
	}
	err = d.GenerateSections()
	if err != nil {
		panic(err.Error())
	}
	numWorkers := 3

	// Create a wait group to wait for all workers to finish
	var wg sync.WaitGroup
	sectionChan := make(chan [3]int)
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			d.worker(ctx, sectionChan, workerID)
		}(i)
	}
	for _, section := range d.Sections {
		sectionChan <- section
	}

	// Close the channel to signal that no more sections will be added
	close(sectionChan)

	// Wait for all workers to finish
	wg.Wait()
	fmt.Println("Download complete!")
	return nil
}

func (d *Download) CalculateSize() error {
	r, err := d.getNewRequest("HEAD")
	if err != nil {
		return fmt.Errorf("read head of file: %v", err)
	}
	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return err
	}
	fmt.Printf("Head request status: %v\n", resp.StatusCode)
	if resp.StatusCode > 299 {
		return errors.New(fmt.Sprintf("Can't process,Head response is %v", resp.StatusCode))
	}

	fileSize, err := strconv.Atoi(resp.Header.Get("Content-Length"))
	if err != nil {
		return err
	}
	d.Size = int(fileSize)
	return nil
}
func (d *Download) GenerateSections() error {

	// s := time.Now()
	if d.FirstSectionSize > d.Size {
		d.FirstSectionSize = d.Size
	}
	// err := d.downloadSection(0, [2]int{0, d.FirstSectionSize - 1})
	// if err != nil {
	// 	return nil, fmt.Errorf("error on calculating speed of first section: %v", err)
	// }
	// duration := time.Now().Sub(s).Seconds()
	// bitrate := CalculateBitrate(d.Size, duration)
	var sections [][3]int

	// _ = bitrate
	step := int(0)
	for i := 0; i < d.Size; i += d.PortionSize {
		// Calculate the size of the current portion
		currentSize := d.PortionSize
		if i+d.PortionSize > d.Size {
			// Adjust the size for the last portion
			currentSize = d.Size - i + 1
		}
		sections = append(sections, [3]int{step, i, i + currentSize - 1})
		step++
	}
	d.Sections = sections
	return nil
}

func CalculateBitrate(size int, downloadDuration float64) string {
	sizeInMB := float64(size / 10000000)
	timeInSeconds := float64(downloadDuration)
	return strconv.FormatFloat(sizeInMB/timeInSeconds, 'f', 4, 64)
}

// Worker function to download sections concurrently
func (d *Download) worker(ctx context.Context, sectionChan <-chan [3]int, workerID int) {
	for {
		select {
		case section, ok := <-sectionChan:
			if !ok {
				// Channel is closed, worker can exit
				return
			}

			// Download the section
			err := d.downloadSection(ctx, section)
			if err != nil {
				fmt.Printf("Worker %v encountered an error: %v\n", workerID, err)
			}

		case <-ctx.Done():
			// Context canceled, worker can exit
			return
		}
	}
}

func (d *Download) Show() {
	fmt.Println(d)
}

// // Start the download
// func (d Download) Do() error {
// 	fmt.Println("Checking URL")
// 	r, err := d.getNewRequest("HEAD")
// 	if err != nil {
// 		return err
// 	}
// 	resp, err := http.DefaultClient.Do(r)
// 	if err != nil {
// 		return err
// 	}
// 	fmt.Printf("Got %v\n", resp.StatusCode)

// 	if resp.StatusCode > 299 {
// 		return errors.New(fmt.Sprintf("Can't process, response is %v", resp.StatusCode))
// 	}

// 	fileSize, err := strconv.Atoi(resp.Header.Get("Content-Length"))
// 	if err != nil {
// 		return err
// 	}
// 	fmt.Printf("Size is %v bytes\n", fileSize)

// 	var sections = make([][2]int, d.TotalSections)
// 	eachSize := fileSize / d.TotalSections
// 	fmt.Printf("Each size is %v bytes\n", eachSize)

// 	// example: if file size is 100 bytes, our section should like:
// 	// [[0 10] [11 21] [22 32] [33 43] [44 54] [55 65] [66 76] [77 87] [88 98] [99 99]]
// 	for i := range sections {
// 		if i == 0 {
// 			// starting byte of first section
// 			sections[i][0] = 0
// 		} else {
// 			// starting byte of other sections
// 			sections[i][0] = sections[i-1][1] + 1
// 		}

// 		if i < d.TotalSections-1 {
// 			// ending byte of other sections
// 			sections[i][1] = sections[i][0] + eachSize
// 		} else {
// 			// ending byte of other sections
// 			sections[i][1] = fileSize - 1
// 		}
// 	}

// 	log.Println(sections)
// 	var wg sync.WaitGroup
// 	// download each section concurrently
// 	for i, s := range sections {
// 		wg.Add(1)
// 		go func(i int, s [2]int) {
// 			defer wg.Done()
// 			err = d.downloadSection(i, s)
// 			if err != nil {
// 				panic(err)
// 			}
// 		}(i, s)
// 	}
// 	wg.Wait()

// 	return d.mergeFiles(sections)
// }
