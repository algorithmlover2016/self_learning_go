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

var (
	DebugLogDir = "debug"
	InfoLogDir  = "info"
	WarnLogDir  = "warn"
	ErrorLogDir = "error"
	FatalLogDir = "fatal"
	PanicLogDir = "panic"
	LogFilename = "pcm.log"

	AccessLogDir      = "access"
	AccessDebugLogDir = "debug"
	AccessInfoLogDir  = "info"
	AccessWarnLogDir  = "warn"
	AccessErrorLogDir = "error"
	AccessFatalLogDir = "fatal"
	AccessPanicLogDir = "panic"
	AccessLogFilename = "access_pcm.log"
	AccessLogger      = log.New()
)

func init() {
	LogDir, err := os.Getwd()
	if err != nil {
		LogDir = "/tmp/"
	}
	LogDir += "/log/"

	AccessLogDir = path.Join(LogDir, AccessLogDir)
	os.MkdirAll(AccessLogDir, 0777)
	AccessDebugLogDir = path.Join(AccessLogDir, AccessDebugLogDir)
	os.MkdirAll(AccessDebugLogDir, 0777)
	AccessInfoLogDir = path.Join(AccessLogDir, AccessInfoLogDir)
	os.MkdirAll(AccessInfoLogDir, 0777)
	AccessWarnLogDir = path.Join(AccessLogDir, AccessWarnLogDir)
	os.MkdirAll(AccessWarnLogDir, 0777)
	AccessErrorLogDir = path.Join(AccessLogDir, AccessErrorLogDir)
	os.MkdirAll(AccessErrorLogDir, 0777)
	AccessFatalLogDir = path.Join(AccessLogDir, AccessFatalLogDir)
	os.MkdirAll(AccessFatalLogDir, 0777)
	AccessPanicLogDir = path.Join(AccessLogDir, AccessPanicLogDir)
	os.MkdirAll(AccessPanicLogDir, 0777)

	DebugLogDir = path.Join(LogDir, DebugLogDir)
	os.MkdirAll(DebugLogDir, 0777)
	InfoLogDir = path.Join(LogDir, InfoLogDir)
	os.MkdirAll(InfoLogDir, 0777)
	WarnLogDir = path.Join(LogDir, WarnLogDir)
	os.MkdirAll(WarnLogDir, 0777)
	ErrorLogDir = path.Join(LogDir, ErrorLogDir)
	os.MkdirAll(ErrorLogDir, 0777)
	FatalLogDir = path.Join(LogDir, FatalLogDir)
	os.MkdirAll(FatalLogDir, 0777)
	PanicLogDir = path.Join(LogDir, PanicLogDir)
	os.MkdirAll(PanicLogDir, 0777)
}

func initAccessLog() {
	AccessLogger.Out = os.Stdout
	accessBaseLogPath := path.Join(AccessLogDir, LogFilename)
	accessWriter, err := rotatelogs.New(
		accessBaseLogPath+".%Y%m%d%H%M",
		rotatelogs.WithLinkName(accessBaseLogPath), // 生成软链，指向最新日志文件
		rotatelogs.WithMaxAge(7*24*time.Hour),      // 文件最大保存时间
		rotatelogs.WithRotationTime(1*time.Hour),   // 日志切割时间间隔
	)
	if err != nil {
		log.Errorf("config access local file system logger error. %v", errors.WithStack(err))
	} else {
		AccessLogger.Out = accessWriter
	}
}

func newLfsHook() log.Hook {
	debugBaseLogPath := path.Join(DebugLogDir, LogFilename)
	infoBaseLogPath := path.Join(InfoLogDir, LogFilename)
	warnBaseLogPath := path.Join(WarnLogDir, LogFilename)
	errorBaseLogPath := path.Join(ErrorLogDir, LogFilename)
	fatalBaseLogPath := path.Join(FatalLogDir, LogFilename)
	panicBaseLogPath := path.Join(PanicLogDir, LogFilename)
	debugWriter, err := rotatelogs.New(
		debugBaseLogPath+".%Y%m%d%H%M",
		rotatelogs.WithLinkName(debugBaseLogPath), // 生成软链，指向最新日志文件
		rotatelogs.WithMaxAge(7*24*time.Hour),     // 文件最大保存时间
		rotatelogs.WithRotationTime(1*time.Hour),  // 日志切割时间间隔
	)
	if err != nil {
		log.Errorf("config local file system logger error. %v", errors.WithStack(err))
	}

	infoWriter, err := rotatelogs.New(
		infoBaseLogPath+".%Y%m%d%H%M",
		rotatelogs.WithLinkName(infoBaseLogPath), // 生成软链，指向最新日志文件
		rotatelogs.WithMaxAge(7*24*time.Hour),    // 文件最大保存时间
		rotatelogs.WithRotationTime(1*time.Hour), // 日志切割时间间隔
	)
	if err != nil {
		log.Errorf("config local file system logger error. %v", errors.WithStack(err))
	}

	warnWriter, err := rotatelogs.New(
		warnBaseLogPath+".%Y%m%d%H%M",
		rotatelogs.WithLinkName(warnBaseLogPath), // 生成软链，指向最新日志文件
		rotatelogs.WithMaxAge(7*24*time.Hour),    // 文件最大保存时间
		rotatelogs.WithRotationTime(1*time.Hour), // 日志切割时间间隔
	)
	if err != nil {
		log.Errorf("config local file system logger error. %v", errors.WithStack(err))
	}

	errorWriter, err := rotatelogs.New(
		errorBaseLogPath+".%Y%m%d%H%M",
		rotatelogs.WithLinkName(errorBaseLogPath), // 生成软链，指向最新日志文件
		rotatelogs.WithMaxAge(7*24*time.Hour),     // 文件最大保存时间
		rotatelogs.WithRotationTime(1*time.Hour),  // 日志切割时间间隔
	)
	if err != nil {
		log.Errorf("config local file system logger error. %v", errors.WithStack(err))
	}

	fatalWriter, err := rotatelogs.New(
		fatalBaseLogPath+".%Y%m%d%H%M",
		rotatelogs.WithLinkName(fatalBaseLogPath), // 生成软链，指向最新日志文件
		rotatelogs.WithMaxAge(7*24*time.Hour),     // 文件最大保存时间
		rotatelogs.WithRotationTime(1*time.Hour),  // 日志切割时间间隔
	)
	if err != nil {
		log.Errorf("config local file system logger error. %v", errors.WithStack(err))
	}

	panicWriter, err := rotatelogs.New(
		panicBaseLogPath+".%Y%m%d%H%M",
		rotatelogs.WithLinkName(panicBaseLogPath), // 生成软链，指向最新日志文件
		rotatelogs.WithMaxAge(7*24*time.Hour),     // 文件最大保存时间
		rotatelogs.WithRotationTime(1*time.Hour),  // 日志切割时间间隔
	)
	if err != nil {
		log.Errorf("config local file system logger error. %v", errors.WithStack(err))
	}

	lfsHook := lfshook.NewHook(lfshook.WriterMap{
		log.DebugLevel: debugWriter, // 为不同级别设置不同的输出目的
		log.InfoLevel:  infoWriter,
		log.WarnLevel:  warnWriter,
		log.ErrorLevel: errorWriter,
		log.FatalLevel: fatalWriter,
		log.PanicLevel: panicWriter,
	}, &log.TextFormatter{DisableColors: true})
	return lfsHook
}

func main() {
	initAccessLog()
	log.AddHook(newLfsHook())
	log.SetLevel(log.DebugLevel)
	log.Info("hello world!")

	AccessLogger.SetLevel(log.InfoLevel)
	AccessLogger.Info("access log")
	AccessLogger.WithFields(log.Fields{
		"animal": "walrus",
		"size":   10,
	}).Info("A group of walrus emerges from the ocean")

	log.WithFields(log.Fields{
		"animal": "walrus",
		"size":   10,
	}).Debug("A group of walrus emerges from the ocean")

	log.WithFields(log.Fields{
		"animal": "walrus",
		"size":   10,
	}).Info("A group of walrus emerges from the ocean")

	log.WithFields(log.Fields{
		"omg":    true,
		"number": 122,
	}).Warn("The group's number increased tremendously!")

	log.WithFields(log.Fields{
		"omg":    true,
		"number": 100,
	}).Fatal("The ice breaks!")

	// log.WithFields(log.Fields{
	// 	"animal": "walrus",
	// 	"size":   10,
	// }).Panic("A group of walrus emerges from the ocean")
}
