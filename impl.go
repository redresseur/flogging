package flogging

import (
	"errors"
	"fmt"
	structure "github.com/redresseur/utils/sturcture"
)

type LoggingServer struct {
	//level type: DEBUG, INFO, WARNING, ERROR, PANIC, FATAL
	//default: info
	level   string

	//If name is nil, os.Stderr will be used as the log sink.
	name    string

	//default: 2006-01-02 15:04:05.000 CST INFO [funcName] "msg info"
	format  string

	//Model of cutting files, "date" or "size"
	//default: date
	model	string

	//Limit the size of each log file
	//This is valid if and only if "size-model" is present
	//default: 5M
	maxFileSize int64

	//Limit the number of log files
	//This is valid if and only if "size-model" is present
	//default: 5
	maxFileNum int

	*FabricLogger
}

type LoggingOption func(log *LoggingServer)

func WithFormat(format string) LoggingOption {
	return func(log *LoggingServer) {
		log.format = format
	}
}

func WithModuleDate() LoggingOption {
	return func(log *LoggingServer) {
		log.model = "date"
	}
}

func WithModuleSize() LoggingOption {
	return func(log *LoggingServer) {
		log.model = "size"
	}
}

func WithMaxFileSize(size int64) LoggingOption {
	return func(log *LoggingServer) {
		log.maxFileSize = size
	}
}

func WithMaxFileNum(num int) LoggingOption {
	return func(log *LoggingServer) {
		log.maxFileNum = num
	}
}

func NewLoggingServer(level string, name string, ops... LoggingOption) *LoggingServer {
	newLogger := &LoggingServer{
		level:  level,
		name:   name,
		format: "%{time:2006-01-02 15:04:05.000 MST} %{level:.4s} [%{shortfunc}] \"%{message}\"",
		model: "date",
		maxFileSize: 1024*1024*5,
		maxFileNum: 10,
	}

	for _, op := range ops{
		op(newLogger)
	}

	return newLogger
}

func (ls *LoggingServer) StartWrite() (err error) {
	data := structure.NewQueue(1)

	var log *Logging
	if ls.name != "" {
		log, err = New(Config{LogSpec:ls.level, Writer:&Writer{data:data}, Format:ls.format})
		if err != nil {
			return err
		}
	} else {
		log, err = New(Config{LogSpec:ls.level, Format:ls.format})
		if err != nil {
			return err
		}
	}
	ls.FabricLogger = log.Logger("log")

	wf := GetWriteFile(data, ls.name)
	switch ls.model {
	case "date":
		wf.StartDateWrite()
	case "size":
		wf.StartSizeWrite(ls.maxFileSize, ls.maxFileNum)
	default:
		return errors.New(fmt.Sprintf("init log faile, model=%s err\n", ls.model))
	}

	return nil
}