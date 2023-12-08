package idm

import (
	"fmt"
	"io"
	"os"
)

// Merge tmp files to single file and delete tmp files
func (d Download) mergeFiles() error {
	f, err := os.OpenFile(fmt.Sprintf("%s%s", d.SavePath, d.FileName), os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.ModePerm)
	if err != nil {
		return err
	}
	defer f.Close()
	for i := 0; i < d.SectionCounts; i++ {
		nSection := d.Sections[i]
		b, err := os.ReadFile(nSection.fileName())
		if err != nil {
			return err
		}
		_, err = f.Write(b)
		if err != nil {
			return err
		}
		err = os.Remove(nSection.fileName())
		if err != nil {
			return err
		}
	}

	return nil
}

type progressReader struct {
	Reader io.Reader
	OnRead func(n int, p []byte)
}

func (pr *progressReader) Read(p []byte) (n int, err error) {
	// Read from the underlying reader
	n, err = pr.Reader.Read(p)

	// Call the OnRead callback
	pr.OnRead(n, p)

	return n, err
}
