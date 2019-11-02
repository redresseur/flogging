package flogging

import (
	"context"
	"os"
)

const (
	DateModel         = "date"
	SizeModel         = "size"
	DefaultMaxSize    = 5 * 1024 * 1024
	DefaultMaxFileNum = 5

	DefaultLogFormat = "%{color}%{time:2006-01-02 15:04:05.000 MST}%{color} " +
		"[%{module}][%{level:.4s}] [%{shortfunc}] \"%{message}\""
	//"%{time:2006-01-02 15:04:05.000 MST} %{level:.4s} [%{shortfunc}] \"%{message}\""
)

type LoggingFactory struct {
	//level type: DEBUG, INFO, WARNING, ERROR, PANIC, FATAL
	//default: info
	level string

	// the root directory of the log files, default /tmp
	rootDir string

	//If name is nil, os.Stderr will be used as the log sink.
	name string

	//default: 2006-01-02 15:04:05.000 CST INFO [funcName] "msg info"
	format string

	//Model of cutting files, "date" or "size"
	//default: date
	model string

	//Limit the size of each log file
	//This is valid if and only if "size-model" is present
	//default: 5M
	maxFileSize int64

	//Limit the number of log files
	//This is valid if and only if "size-model" is present
	//default: 5
	maxFileNum int

	log *Logging

	ctx context.Context
}

type LoggingOption func(log *LoggingFactory)

func WithFormat(format string) LoggingOption {
	return func(log *LoggingFactory) {
		log.format = format
	}
}

func WithRootDir(rootDir string) LoggingOption {
	return func(log *LoggingFactory) {
		log.rootDir = rootDir
	}
}

func WithModuleDate() LoggingOption {
	return func(log *LoggingFactory) {
		log.model = DateModel
	}
}

func WithModuleSize() LoggingOption {
	return func(log *LoggingFactory) {
		log.model = SizeModel
	}
}

func WithMaxFileSize(size int64) LoggingOption {
	return func(log *LoggingFactory) {
		log.maxFileSize = size
	}
}

func WithMaxFileNum(num int) LoggingOption {
	return func(log *LoggingFactory) {
		log.maxFileNum = num
	}
}

func NewLoggingFactory(level string, name string, ops ...LoggingOption) *LoggingFactory {
	newLogger := &LoggingFactory{
		level:       level,
		name:        name,
		format:      DefaultLogFormat,
		model:       DateModel,
		maxFileSize: DefaultMaxSize,
		maxFileNum:  DefaultMaxFileNum,
		rootDir:     os.TempDir(),
		ctx:         context.Background(),
	}

	for _, op := range ops {
		op(newLogger)
	}

	return newLogger
}

func (ls *LoggingFactory) Logger(name string) *FabricLogger {
	if ls.log != nil {
		return ls.log.Logger(name)
	}

	return nil
}

func (ls *LoggingFactory) Initial() (err error) {
	if ls.name != "" {
		w, err := NewWriter(ls.ctx, &WriterConfig{
			Dir:          ls.rootDir,
			Prefix:       ls.name,
			Model:        ls.model,
			MaxSize:      ls.maxFileSize,
			MaxFileCount: ls.maxFileNum,
		})
		ls.log, err = New(Config{LogSpec: ls.level, Writer: w, Format: ls.format})
		if err != nil {
			return err
		}
	} else {
		ls.log, err = New(Config{LogSpec: ls.level, Format: ls.format})
		if err != nil {
			return err
		}
	}

	return nil
}
