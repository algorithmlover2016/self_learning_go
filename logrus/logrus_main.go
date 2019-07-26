package main

import (
	"flag"
	"fmt"
	"github.com/lestrrat/go-file-rotatelogs"
	"github.com/pkg/errors"
	"github.com/rifflock/lfshook"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	DebugLogDir string = "debug"
	InfoLogDir  string = "info"
	WarnLogDir  string = "warn"
	ErrorLogDir string = "error"
	FatalLogDir string = "fatal"
	PanicLogDir string = "panic"

	AccessLogDir string = "access"
	FieldKeyMsg  string = log.FieldKeyMsg
)

type Fields = log.Fields
type Name = string

var (
	AccessLogger = log.New()
	Log          = log.New()

	// configuration logs
	LogDirPtr = flag.String("log_dir", "", "If non-empty, write log files in this directory")
	pid       = os.Getpid()
	program   = filepath.Base(os.Args[0])
	host      = "unknownhost"
	userName  = "unknownuser"
)

func print(msg string) {
	fmt.Println(msg)
}

// shortHostname returns its argument, truncating at the first period.
// For instance, given "www.google.com" it returns "www".
func shortHostname(hostname string) string {
	if i := strings.Index(hostname, "."); i >= 0 {
		return hostname[:i]
	}
	return hostname
}

// logName returns a new log file name containing tag, with start time t, and
// the name for the symlink for tag.
func logName(logLevelTag string) (name string) {
	name = fmt.Sprintf("%s.%s.%s.log.%s.%d",
		program,
		host,
		userName,
		logLevelTag,
		pid)
	return
}

func init() {
	flag.Parse()

	workDir, err := os.Getwd()
	if err != nil {
		workDir = ""
	} else {
		workDir += "/"
	}
	AccessLogger.SetOutput(ioutil.Discard)
	Log.SetOutput(ioutil.Discard)

	h, err := os.Hostname()
	if err == nil {
		host = shortHostname(h)
	}

	current, err := user.Current()
	if err == nil {
		userName = current.Username
	}

	// Sanitize userName since it may contain filepath separators on Windows.
	userName = strings.Replace(userName, `\`, "_", -1)
	LogDir := *LogDirPtr

	if len(LogDir) == 0 || strings.HasPrefix(LogDir, ".") {
		var err error
		LogDir, err = os.Getwd()
		if err != nil {
			LogDir = "/tmp/"
		}
		LogDir += "/log/"
	}

	// create log folder
	for _, level := range log.AllLevels {
		levelStr := level.String()
		levelLogDir := path.Join(LogDir, levelStr)
		os.MkdirAll(levelLogDir, 0777)
		accessLevelLogDir := path.Join(LogDir, AccessLogDir, levelStr)
		os.MkdirAll(accessLevelLogDir, 0777)
	}
	// add log Hook
	AccessLogger.AddHook(newLfsHook(path.Join(LogDir, AccessLogDir), true, workDir))
	// set logLevel
	AccessLogger.SetLevel(log.TraceLevel)
	AccessLogger.SetReportCaller(true)
	Log.AddHook(newLfsHook(LogDir, false, workDir))
	Log.SetLevel(log.TraceLevel)
	Log.SetReportCaller(true)
}

func newLfsHook(logDir string, color bool, workDir string) log.Hook {
	// make a map to save the logLevel -> rotatelogs.RotateLogs object
	logPathWriterMap := make(map[string]*rotatelogs.RotateLogs)
	for _, level := range log.AllLevels {
		// change the type of log level to string
		levelStr := level.String()
		upperLevelStr := strings.ToUpper(levelStr)
		// get the destination of log path
		levelLogPath := path.Join(logDir, levelStr,
			fmt.Sprintf("%s.%s", program, upperLevelStr))

		rotateLogPath := path.Join(logDir, levelStr,
			fmt.Sprintf("%s.%s.%s.log.%s.%%Y%%m%%d-%%H%%M%%S.%d",
				// program, upperLevelStr, host, userName, 233))
				program, upperLevelStr, host, userName, pid))

		// create the rotatelogs object
		Writer, err := rotatelogs.New(
			rotateLogPath,
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
	}, &log.TextFormatter{
		DisableColors:          color,
		DisableLevelTruncation: true,
		TimestampFormat:        "2006-01-02 15:04:05.000",
		// if you donot set CallerPrettyfier, the result is funcname, fmt.Sprintf("%s:%d", f.File, f.line)
		CallerPrettyfier: func(f *runtime.Frame) (string, string) {
			s := strings.Split(f.Function, ".")
			funcname := s[len(s)-1]
			// _, filename := path.Split(f.File)
			filename := (strings.TrimPrefix(f.File, workDir))
			filename = fmt.Sprintf("%s:%d", filename, f.Line)
			return funcname, filename
		},
		SortingFunc: func(keyLists []string) {
			for index, val := range keyLists {
				if FieldKeyMsg == val && len(keyLists) > (index+1) {
					keyLists = append(keyLists[:index], keyLists[index+1:]...)
					keyLists = append(keyLists, val)
					break
				}
			}
		},
	})
	return lfsHook
}

func main() {
	print(FieldKeyMsg)
	var aliasName Name = "fafsadfsda"
	print(aliasName)
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

	Log.WithFields(Fields{
		"omg":    true,
		"number": 100,
	}).Fatal("The ice breaks!")

	// log.WithFields(log.Fields{
	// 	"animal": "walrus",
	// 	"size":   10,
	// }).Panic("A group of walrus emerges from the ocean")
}
