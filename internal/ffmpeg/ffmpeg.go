package ffmpeg

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"recorder/config"
	"recorder/internal/kvm"
	"recorder/pkg/logger"
	"time"
)

func init() {
	if _, err := os.Stat("/usr/bin/ffmpeg"); os.IsNotExist(err) {
		fmt.Println("ffmpeg not found")
	}
}

func Record(ch chan<- string, mh *kvm.Kvm, ctx context.Context) {
	hostname := mh.Hostname
	url := mh.Stream_url
	video_path := config.Viper.GetString("RECORDING_PATH") + hostname + "/"
	image_path := config.Viper.GetString("IMAGE_PATH") + hostname + "/"
	err := os.RemoveAll(video_path)
	if err != nil {
		logger.Error(err.Error())
	}
	if _, err := os.Stat(video_path); os.IsNotExist(err) {
		err := os.Mkdir(video_path, 0777)
		if err != nil {
			logger.Error(err.Error())
		}
		// TODO: handle error
	}
	err = os.RemoveAll(image_path)
	if err != nil {
		logger.Error(err.Error())
	}
	if _, err := os.Stat(image_path); os.IsNotExist(err) {
		err := os.Mkdir(image_path, 0777)
		if err != nil {
			logger.Error(err.Error())
		}
		// TODO: handle error
	}
	cmd := exec.Command("ffmpeg", "-loglevel", "quiet","-y", "-i", url,
		"-codec", "libx264", "-preset", "ultrafast", "-f", "hls", "-strftime", "1" ,"-hls_segment_filename", video_path+"%Y-%m-%d_%H-%M-%S.ts", video_path+"all.m3u8",
		"-r", "0.2", "-update", "1", image_path+hostname+".png")
	logger.Info(cmd.String())
	in, err := cmd.StdinPipe()
	if err != nil {
		logger.Error(err.Error())
	}
	_, err = cmd.StderrPipe()
	if err != nil {
		logger.Error(err.Error())
	}
	// logger.Info("test1")
	err = cmd.Start()
	mh.Start_record_time = time.Now().Unix()
	if err != nil {
		logger.Error(err.Error())
	}
	cmdDone := make(chan error)
	go func() {
		cmdDone <- cmd.Wait()
	}()
	select {
	case <-ctx.Done():
		fmt.Println("send exit signal")
		io.WriteString(in, "q")
		if err != nil {
			logger.Error(err.Error())
		}
	case err := <-cmdDone:
		if err != nil {
			logger.Error(err.Error())
		}
	}
	ch <- hostname
}
