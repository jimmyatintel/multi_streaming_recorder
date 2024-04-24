package ai

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"image"
	"image/png"
	"io"
	"os"
	"recorder/config"
	"recorder/internal/cropping"
	"recorder/internal/logpicqueue"
	"recorder/internal/ssim"
	"recorder/internal/structure"
	"recorder/pkg/fileoperation"
	"recorder/pkg/logger"
	dut_query "recorder/pkg/mariadb/dut"
	"recorder/pkg/rabbitmq"
	"strconv"
	"time"
)

var AI_list []string

type Message struct {
	Hostname     string    `json:"hostname"`
	Machine_name string    `json:"machine_name"`
	Image        string    `json:"image"`
	Coord        []float64 `json:"coord"`
	Locked       int       `json:"locked"`
	// Path     string `json:"path"`
}

func Start_ai_monitoring(ctx context.Context) {
	_, err := rabbitmq.Declare("AI_queue1")
	if err != nil {
		logger.Error("Declare to rabbit error: " + err.Error())
		return
	}
	go FS_monitor_ramdisk(ctx)

	<-ctx.Done()

}

func Process_AI_result(hostname string, machine_name string) {
	// unit := unit_query.Get_unitbyhostname(hostname)
	// sta := dut_query.Get_dut_status(unit.Machine_name)
	Ai_result := dut_query.Get_AI_result(machine_name)
	if Ai_result.Hostname == "null" {
		logger.Error("Machine " + machine_name + " not found in database")
		return
	}
	if len(Ai_result.Coords) == 0 {
		return
	}
	slow_path := config.Viper.GetString("slow_path")
	cropped_path := config.Viper.GetString("cropped_path")
	cropping.Switch_picture_if_exist(cropped_path + hostname + "_cropped.png")
	var cropped_image image.Image
	var err error
	cropped_image, err = cropping.Crop_image(slow_path+hostname+".png", Ai_result.Coords, cropped_path+hostname+"_cropped.png")
	if err != nil {
		logger.Error(err.Error())
	}
	err = logpicqueue.SendtoLogPicChannel(hostname, cropped_image)
	if err != nil {
		logger.Error(err.Error())
	}
	if Ai_result.Label == 0 {
		dut_query.Update_dut_status(machine_name, 0)
		dut_query.Update_dut_cnt(machine_name, 0)
	} else {
		dut_info := dut_query.Get_dut_status(machine_name)
		if !fileoperation.FileExists(cropped_path + hostname + "_cropped_old.png") {
			return
		}
		ssim_result, err := ssim.Ssim_cal(cropped_path+hostname+"_cropped.png", cropped_path+hostname+"_cropped_old.png")
		if err != nil {
			logger.Error(err.Error())
			return
		}
		if ssim_result >= dut_info.Ssim {
			dut_query.Update_dut_cnt(machine_name, dut_info.Cycle_cnt+1)
			dut_info.Cycle_cnt++
		} else {
			dut_query.Update_dut_cnt(machine_name, 0)
		}
		if dut_info.Cycle_cnt == dut_info.Threshhold {
			// dut_query.Update_dut_status(hostname, 4)
			freeze_process(machine_name)
		}
		logger.Debug("SSIM result: " + strconv.FormatFloat(ssim_result, 'f', 6, 64))
	}
	if Ai_result.Label == 2 {
		//todo: handle restart type
	}
}
func freeze_process(machine_name string) {
	copyFileFromQueue(machine_name)

}
func copyFileFromQueue(machine_name string) {
	logpicqueue.BlockLogPicChannel(machine_name)
	defer logpicqueue.UnblockLogPicChannel(machine_name)
	var index = 0
	for {
		image := logpicqueue.GetChannelContent(machine_name)
		if image == nil {
			break
		}
		fileoperation.CreateFolderifNotExist(config.Viper.GetString("logimage_path") + machine_name)
		outputImagePath := config.Viper.GetString("logimage_path") + machine_name + "/" + strconv.Itoa(index) + ".png"
		outputFile, err := os.Create(outputImagePath)
		if err != nil {
			logger.Error(err.Error())
			return
		}
		err = png.Encode(outputFile, image)
		if err != nil {
			logger.Error(err.Error())
			return
		}
		index++
	}

}
func Send_to_rabbitMQ(hostname string, machine_name string, locked string, path string, expire_time string) (err error) {
	var message Message
	message.Hostname = hostname
	message.Machine_name = machine_name
	message.Locked = 0
	if locked != "" {
		message.Locked = 1
		message.Coord = structure.Coord_s2f(locked)
	}
	time.Sleep(100 * time.Millisecond)
	// logger.Info(path)
	imageFile, err := os.Open(path)
	if err != nil {
		return err
	}
	imageData, err := io.ReadAll(imageFile)
	if err != nil {
		return err
	}
	message.Image = base64.StdEncoding.EncodeToString(imageData)
	jsonMessage, _ := json.Marshal(message)
	rabbitmq.Publish_with_expiration("AI_queue1", jsonMessage, expire_time)
	imageFile.Close()
	return nil
}
