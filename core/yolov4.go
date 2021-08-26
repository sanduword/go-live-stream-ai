package core

import (
	"bufio"
	"fmt"
	"image"
	"image/color"
	"os"
	"sort"
	"strconv"

	"github.com/pfhds/live-stream-ai/models"
	"gocv.io/x/gocv"
)

var YoloMs *YoloManager

// 读取label
func readLabels(path string) []string {
	var classes []string
	read, _ := os.Open(fmt.Sprintf("%v/coco.names", path))
	defer read.Close()
	scanner := bufio.NewScanner(read)
	for scanner.Scan() {
		classes = append(classes, scanner.Text())
	}
	return classes
}

// getOutputsNames : YOLO Layer
func getOutputsNames(net *gocv.Net) []string {
	var outputLayers []string
	for _, i := range net.GetUnconnectedOutLayers() {
		layer := net.GetLayer(i)
		layerName := layer.GetName()
		if layerName != "_input" {
			outputLayers = append(outputLayers, layerName)
		}
	}

	return outputLayers
}

// 启动yolov4模型初始化工作
func (yolo *YoloManager) StartYolo(config *models.ServerConfig) {
	YoloMs = &YoloManager{}
	YoloMs.Labels = readLabels(config.YoloPath)
	net := gocv.ReadNet(fmt.Sprintf("%v/yolov4.weights", config.YoloPath), fmt.Sprintf("%v//yolov4.cfg", config.YoloPath))
	//defer net.Close()
	//gocv.NetBackendCUDA
	net.SetPreferableBackend(gocv.NetBackendType(gocv.NetBackendCUDA))
	// gocv.NetTargetCUDA
	net.SetPreferableTarget(gocv.NetTargetType(gocv.NetTargetCUDA))
	YoloMs.Net = &net

	YoloMs.OutputNames = getOutputsNames(&net)
	YoloMs.ScoreThreshold = config.Score
	YoloMs.NmsThreshold = config.Nms
	YoloMs.LabelData = make(map[string]*LabelResult)
}

// PostProcess : All Detect Box
func PostProcess(frame gocv.Mat, outs *[]gocv.Mat) ([]image.Rectangle, []float32, []int) {
	var classIds []int
	var confidences []float32
	var boxes []image.Rectangle
	for _, out := range *outs {

		data, _ := out.DataPtrFloat32()
		for i := 0; i < out.Rows(); i, data = i+1, data[out.Cols():] {

			scoresCol := out.RowRange(i, i+1)

			scores := scoresCol.ColRange(5, out.Cols())
			_, confidence, _, classIDPoint := gocv.MinMaxLoc(scores)
			if confidence > 0.5 {

				centerX := int(data[0] * float32(frame.Cols()))
				centerY := int(data[1] * float32(frame.Rows()))
				width := int(data[2] * float32(frame.Cols()))
				height := int(data[3] * float32(frame.Rows()))

				left := centerX - width/2
				top := centerY - height/2
				classIds = append(classIds, classIDPoint.X)
				confidences = append(confidences, float32(confidence))
				boxes = append(boxes, image.Rect(left, top, width, height))
			}
		}
	}
	return boxes, confidences, classIds
}

// drawRect : Detect Class to Draw Rect
func drawRect(img gocv.Mat, boxes []image.Rectangle, classes []string, classIds []int, indices []int, confidences []float32) (gocv.Mat, map[string]LabelResult) {
	detectClass := make(map[string]LabelResult)
	for _, idx := range indices {
		if idx == 0 {
			continue
		}
		labelName := classes[classIds[idx]] + " " + strconv.FormatFloat(float64(confidences[idx]), 'f', 2, 32)
		sizeConfig := gocv.GetTextSize(labelName, gocv.FontHersheyComplexSmall, 1.2, 2)
		x0 := boxes[idx].Max.X
		y0 := boxes[idx].Max.Y
		x1 := boxes[idx].Max.X + boxes[idx].Min.X
		y1 := boxes[idx].Max.Y + boxes[idx].Min.Y
		if x0+2 >= img.Cols() {
			x1 = img.Cols() - 1*2
		}
		if y1+2 >= img.Rows() {
			y1 = img.Rows() - 1*2
		}
		// 画框
		gocv.Rectangle(
			&img,
			image.Rect(x0, y0, x1, y1),
			color.RGBA{255, 0, 191, 0},
			1)
		// 文字填充
		gocv.Rectangle(
			&img,
			image.Rect(x0, y0-sizeConfig.Y, x0+sizeConfig.X, y0),
			color.RGBA{255, 0, 191, 0},
			-1)
		// 文字
		gocv.PutText(
			&img,
			labelName,
			image.Point{x0, y0 - 3}, gocv.FontHersheyComplexSmall,
			1.2,
			color.RGBA{255, 255, 255, 0},
			2)
		model, ok := detectClass[classes[classIds[idx]]]
		if ok {
			model.Warn = model.Warn + 1
		} else {
			model = LabelResult{
				LabelName: classes[classIds[idx]],
				Warn:      1,
			}
		}
		detectClass[classes[classIds[idx]]] = model
		//detectClass = append(detectClass, labelName)
	}
	return img, detectClass
}

// 计算识别的结果集
func GetWarnData(warnData map[string]LabelResult) []*LabelResult {
	if len(warnData) == 0 {
		return nil
	}
	for k, v := range warnData {
		model, ok := YoloMs.LabelData[k]
		if ok {
			model.Warn = model.Warn + v.Warn
		} else {
			model = &LabelResult{
				LabelName: v.LabelName,
				Warn:      v.Warn,
			}
		}

		YoloMs.LabelData[k] = model
	}

	var warnResult []*LabelResult
	for _, v := range YoloMs.LabelData {
		warnResult = append(warnResult, v)
	}
	sort.Slice(warnResult, func(i, j int) bool {
		return warnResult[i].Warn > warnResult[j].Warn
	})
	return warnResult
}

// 处理单张图像
func RunDetectImg(src gocv.Mat) (gocv.Mat, []*LabelResult) {
	img := src.Clone()
	img.ConvertTo(&img, gocv.MatTypeCV32F)
	// 默认为416，降低这个值可以在CPU架构上提高识别速度，但是会降低识别精度
	blob := gocv.BlobFromImage(img, 1/255.0, image.Pt(128, 128), gocv.NewScalar(0, 0, 0, 0), true, false)
	YoloMs.Net.SetInput(blob, "")
	probs := YoloMs.Net.ForwardLayers(YoloMs.OutputNames)
	boxes, confidences, classIds := PostProcess(img, &probs)
	img.Close()
	blob.Close()
	indices := make([]int, len(YoloMs.Labels))
	if len(boxes) == 0 { // No Classes
		return src, nil
	}

	gocv.NMSBoxes(boxes, confidences, YoloMs.ScoreThreshold, YoloMs.NmsThreshold, indices)

	img2, warnData := drawRect(src, boxes, YoloMs.Labels, classIds, indices, confidences)
	warnResult := GetWarnData(warnData)
	return img2, warnResult
}
