package main

import (
	"fmt"
	"go-download-manager/pkg/idm"
)

type Idm struct {
	SavePath string
}

type Job struct {
}

func main() {
	d, err := idm.NewDownload(
		// fmt.Sprintf("https://dl2.soft98.ir/soft/i/Internet.Download.Manager.6.42.Build.1.Retail.zip?1700930984"),
		fmt.Sprintf("https://dl2.soft98.ir/soft/m/Mozilla.Firefox.120.0.EN.x64.zip?1701093556"),
		5000,
		2000000,
		5,
	)
	if err != nil {
		panic(err.Error())
	}
	d.Start()

	d.Show()
}
