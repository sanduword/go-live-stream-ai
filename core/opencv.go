package core

import (
	"encoding/base64"
	"image"
	"time"

	"github.com/pfhds/live-stream-ai/utils/log"
	"gocv.io/x/gocv"
)

// 启动识别任务
func (c *Client) StartOpencvStream() {
	var err error
	c.OpenCv, err = gocv.VideoCaptureFile(c.Url)

	if err != nil {
		log.Errorf("open stream url error:%v", err)
		return
	}
	//defer webcam.Close()
	defer func() {
		c.OpenCv.Close()
		close(c.BaseImg)
		close(c.WarnData)
	}()

	//width := strconv.FormatFloat(c.openCv.Get(gocv.VideoCaptureFrameWidth), 'f', -1, 64)
	//c.openCv.Set(gocv.VideoCaptureFrameWidth, 640)
	//c.openCv.Set(gocv.VideoCaptureFrameHeight, 480)
	//height := strconv.FormatFloat(c.openCv.Get(gocv.VideoCaptureFrameWidth), 'f', -1, 64)

	img := gocv.NewMat()
	defer img.Close()

	no := 0
	var trsImg = gocv.NewMat()
	defer trsImg.Close()
	for {
		if ok := c.OpenCv.Read(&img); !ok {
			log.Errorln("open device error")
			break
		}
		if img.Empty() {
			continue
		}

		// 抽帧
		if no%10 == 0 {
			gocv.Resize(img, &trsImg, image.Point{X: 640, Y: 480}, 0, 0, gocv.InterpolationDefault)
			detectImg, warnData := RunDetectImg(trsImg.Clone())
			// 加入高斯滤波模糊处理关键信息
			gsmh := detectImg.RowRange(detectImg.Rows()-30, detectImg.Rows())
			gocv.GaussianBlur(gsmh, &gsmh, image.Pt(31, 31), 0, 0, gocv.BorderDefault)
			data, err := gocv.IMEncode(".jpg", detectImg)
			if err != nil {
				log.Errorf("encode, error: %v", err)
			} else {
				n := base64.StdEncoding.EncodedLen(len(data.GetBytes()))
				dst := make([]byte, n)
				base64.StdEncoding.Encode(dst, data.GetBytes())
				urldata := "data:image/jpeg;base64," + string(dst)
				formimg := []byte(urldata)
				c.BaseImg <- formimg
				if len(warnData) > 0 {
					c.WarnData <- warnData
				}

				log.Infof("正在处理第：%v 帧图像", no)
			}

			detectImg.Close()
			data.Close()

			time.Sleep(time.Millisecond * time.Duration(50))

		}
		no++
	}
}
