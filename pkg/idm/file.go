package idm

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

// Merge tmp files to single file and delete tmp files
func (d Download) mergeFiles(sections []section) error {
	f, err := os.OpenFile(fmt.Sprintf("%s%s", d.SavePath, d.FileName), os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.ModePerm)
	if err != nil {
		return err
	}
	defer f.Close()
	for i := range sections {
		tmpFileName := fmt.Sprintf("%s-%v.tmp", d.FileName, i)
		b, err := ioutil.ReadFile(tmpFileName)
		if err != nil {
			return err
		}
		fmt.Println(b)
		n, err := f.Write(b)
		if err != nil {
			return err
		}
		err = os.Remove(tmpFileName)
		if err != nil {
			return err
		}
		fmt.Printf("%v bytes merged\n", n)
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
