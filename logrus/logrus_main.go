package main

import (
	"github.com/lestrrat/go-file-rotatelogs"
	"github.com/pkg/errors"
	"github.com/rifflock/lfshook"
	log "github.com/sirupsen/logrus"
	"os"
	"path"
	"time"
)

const (
	DebugLogDir string = "debug"
	InfoLogDir  string = "info"
	WarnLogDir  string = "warn"
	ErrorLogDir string = "error"
	FatalLogDir string = "fatal"
	PanicLogDir string = "panic"

	LogFilename string = "pcm.log"

	AccessLogDir      string = "access"
	AccessLogFilename string = "access_pcm.log"
)

var (
	AccessLogger = log.New()
	Log          = log.New()
)

func init() {

	LogDir, err := os.Getwd()
	if err != nil {
		LogDir = "/tmp/"
	}

	LogDir += "/log/"

	// create log folder
	for _, level := range log.AllLevels {
		levelStr := level.String()
		levelLogDir := path.Join(LogDir, levelStr)
		os.MkdirAll(levelLogDir, 0777)
		accessLevelLogDir := path.Join(LogDir, AccessLogDir, levelStr)
		os.MkdirAll(accessLevelLogDir, 0777)
	}
	// add log Hook
	AccessLogger.AddHook(newLfsHook(path.Join(LogDir, AccessLogDir), AccessLogFilename, true))
	// set logLevel
	AccessLogger.SetLevel(log.TraceLevel)
	Log.AddHook(newLfsHook(LogDir, LogFilename, false))
	Log.SetLevel(log.TraceLevel)
}

func newLfsHook(logDir string, logFile string, color bool) log.Hook {
	// make a map to save the logLevel -> rotatelogs.RotateLogs object
	logPathWriterMap := make(map[string]*rotatelogs.RotateLogs)
	for _, level := range log.AllLevels {
		// change the type of log level to string
		levelStr := level.String()
		// get the destination of log path
		levelLogPath := path.Join(logDir, levelStr, logFile)
		// create the rotatelogs object
		Writer, err := rotatelogs.New(
			levelLogPath+".%Y%m%d_%H%M%S",
			rotatelogs.WithLinkName(levelLogPath),    // 生成软链，指向最新日志文件
			rotatelogs.WithMaxAge(7*24*time.Hour),    // 文件最大保存时间
			rotatelogs.WithRotationTime(1*time.Hour), // 日志切割时间间隔
		)
		if err != nil {
			log.Errorf("config %s local file system logger error. %v", levelLogPath, errors.WithStack(err))
			logPathWriterMap[levelStr] = nil
		}
		logPathWriterMap[levelStr] = Writer
	}

	lfsHook := lfshook.NewHook(lfshook.WriterMap{
		log.TraceLevel: logPathWriterMap[log.TraceLevel.String()], // 为不同级别设置不同的输出目的
		log.DebugLevel: logPathWriterMap[log.DebugLevel.String()], // 为不同级别设置不同的输出目的
		log.InfoLevel:  logPathWriterMap[log.InfoLevel.String()],
		log.WarnLevel:  logPathWriterMap[log.WarnLevel.String()],
		log.ErrorLevel: logPathWriterMap[log.ErrorLevel.String()],
		log.FatalLevel: logPathWriterMap[log.FatalLevel.String()],
		log.PanicLevel: logPathWriterMap[log.PanicLevel.String()],
	}, &log.TextFormatter{DisableColors: color, DisableLevelTruncation: true,
		TimestampFormat: "2006-01-02 15:04:05.000"})
	return lfsHook
}

func main() {
	Log.Info("hello world!")

	AccessLogger.Info("access log")

	AccessLogger.WithFields(log.Fields{
		"animal": "walrus",
		"size":   10,
	}).Trace("A group of walrus emerges from the ocean")

	AccessLogger.WithFields(log.Fields{
		"animal": "walrus",
		"size":   10,
	}).Debug("A group of walrus emerges from the ocean")

	AccessLogger.WithFields(log.Fields{
		"animal": "walrus",
		"size":   10,
	}).Info("A group of walrus emerges from the ocean")

	AccessLogger.WithFields(log.Fields{
		"animal": "walrus",
		"size":   10,
	}).Warn("A group of walrus emerges from the ocean")

	AccessLogger.WithFields(log.Fields{
		"animal": "walrus",
		"size":   10,
	}).Error("A group of walrus emerges from the ocean")

	Log.WithFields(log.Fields{
		"omg":    true,
		"number": 122,
	}).Trace("The group's number increased tremendously!")

	Log.WithFields(log.Fields{
		"omg":    true,
		"number": 122,
	}).Debug("The group's number increased tremendously!")

	Log.WithFields(log.Fields{
		"omg":    true,
		"number": 122,
	}).Info("The group's number increased tremendously!")

	Log.WithFields(log.Fields{
		"omg":    true,
		"number": 122,
	}).Warn("The group's number increased tremendously!")

	Log.WithFields(log.Fields{
		"omg":    true,
		"number": 100,
	}).Fatal("The ice breaks!")

	// log.WithFields(log.Fields{
	// 	"animal": "walrus",
	// 	"size":   10,
	// }).Panic("A group of walrus emerges from the ocean")
}
