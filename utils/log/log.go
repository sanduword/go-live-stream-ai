package log

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/gin-gonic/gin"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/pfhds/live-stream-ai/utils/stream"
	"github.com/pfhds/live-stream-ai/utils/xtime"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
)

var g *logrus.Logger

func Init() {
	g = New(os.Stdout)
}

func New(out io.Writer) *logrus.Logger {
	log := logrus.New()
	log.Out = out
	log.SetLevel(logrus.DebugLevel)
	log.SetFormatter(&logrus.TextFormatter{
		TimestampFormat: xtime.TimeFullFormat,
		DisableColors:   true,
	})
	return log
}

// 本地输出
func NewLocal(logPath string) gin.HandlerFunc {
	logFilePath := logPath
	err := stream.CreateMoreFolder(logFilePath)
	if err != nil {
		Errorf("create folder error:%v \n", err)
	}
	logFileName := time.Now().Format(xtime.TimeShortFormat) + ".log"
	// 日志文件
	fileName := path.Join(logFilePath, logFileName)

	// 写入文件
	src, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		Errorf("open file：%v error:%v \n", fileName, err)
	}

	log := New(src)

	// 设置 rotatelogs
	logWriter, _ := rotatelogs.New(
		// 分割后的文件名称
		fileName+".%Y%m%d.log",

		// 生成软链，指向最新日志文件
		rotatelogs.WithLinkName(fileName),

		// 设置最大保存时间(7天)
		rotatelogs.WithMaxAge(7*24*time.Hour),

		// 设置日志切割时间间隔(1天)
		rotatelogs.WithRotationTime(24*time.Hour),
	)

	writeMap := lfshook.WriterMap{
		logrus.InfoLevel:  logWriter,
		logrus.FatalLevel: logWriter,
		logrus.DebugLevel: logWriter,
		logrus.WarnLevel:  logWriter,
		logrus.ErrorLevel: logWriter,
		logrus.PanicLevel: logWriter,
	}

	lfHook := lfshook.NewHook(writeMap, &logrus.JSONFormatter{})

	// 新增 Hook
	log.AddHook(lfHook)

	return logger(log)
}

// 控制台显示
func NewLogConsole() gin.HandlerFunc {
	log := New(os.Stdout)

	return logger(log)
}

func logger(logger logrus.FieldLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 开始时间
		startTime := time.Now()
		// 处理请求
		c.Next()
		// 结束时间
		endTime := time.Now()
		// 执行时间
		latencyTime := endTime.Sub(startTime)
		// 请求方式
		reqMethod := c.Request.Method
		// 请求路由
		reqUri := c.Request.RequestURI
		// 状态码
		statusCode := c.Writer.Status()
		// 请求IP
		clientIP := c.ClientIP()

		entry := logger.WithFields(logrus.Fields{
			"statusCode":  statusCode,
			"latencyTime": latencyTime,
			"clientIP":    clientIP,
			"reqMethod":   reqMethod,
			"reqUri":      reqUri,
		})

		if len(c.Errors) > 0 {
			entry.Error(c.Errors.ByType(gin.ErrorTypePrivate).String())
		} else {
			msg := fmt.Sprintf("| %3d | %13v | %15s | %s | %s |",
				statusCode,
				latencyTime,
				clientIP,
				reqMethod,
				reqUri)
			if statusCode >= http.StatusInternalServerError {
				entry.Error(msg)
			} else if statusCode >= http.StatusBadRequest {
				entry.Warn(msg)
			} else {
				entry.Info(msg)
			}
		}
	}
}

func Info(args ...interface{}) {
	g.Info(args)
}

func Infoln(args ...interface{}) {
	g.Infoln(args)
}

func Infof(format string, args ...interface{}) {
	g.Infof(format, args)
}

func Warn(format string, args ...interface{}) {
	g.Warn(format, args)
}

func Error(format string, args ...interface{}) {
	g.Error(format, args)
}

func Errorf(format string, args ...interface{}) {
	g.Errorf(format, args)
}

func Fatalf(format string, args ...interface{}) {
	g.Fatalf(format, args)
}

func Errorln(format string, args ...interface{}) {
	g.Errorln(format, args)
}
