package main

import (
	"context"
	"fmt"
	"go-download-manager/pkg/idm"
)

type Idm struct {
	SavePath string
}

type Job struct {
}

// TODO: implimrent cheksum check after asmebeling file
func main() {
	d, err := idm.NewDownload(
		"https://dl.sakhamusic.ir/94/aban/03/Moshen%20Namjoo%20&%20Nerve%20-%20Hafiz%20[128].mp3",
		5,
		1,
	)
	if err != nil {
		panic(err.Error())
	}

	ctx, cancel := context.WithCancel(context.Background())
	_ = cancel
	d.Start(ctx)
	// func() {
	// 	time.Sleep(time.Second * 5)
	// 	cancel()
	// }()
	fmt.Println("here")
	d.Show()
}
