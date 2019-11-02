package flogging

import (
	"testing"
)

func TestNewLoggingFactory(t *testing.T) {
	//1. Create a new log server
	logging := NewLoggingFactory("debug", "flogging",
		WithModuleDate(), WithMaxFileSize(1024),
		WithMaxFileNum(5), WithRootDir("./tmp"))
	//2. Start writing files
	err := logging.Initial()
	if err != nil {
		t.Fatalf("init log faile, %v\n", err)
	}

	logger = logging.Logger("test")
	//3. Log writing
	for i := 0; i < 10240; i++ {
		logger.Debug("test debug ------------------", i)
	}

	logger.Sync()
}

func BenchmarkLoggingFactory_Logger(b *testing.B) {
	once.Do(func() {
		//1. Create a new log server
		logging := NewLoggingFactory("debug", "flogging",
			WithModuleDate(), WithMaxFileSize(1024),
			WithMaxFileNum(5), WithRootDir("./tmp"))
		//2. Start writing files
		err := logging.Initial()
		if err != nil {
			b.Fatalf("init log faile, %v\n", err)
		}

		logger = logging.Logger("test")
	})

	//3. Log writing
	for i := 0; i < b.N; i++ {
		logger.Debug("test debug ------------------", i)
	}

	once.Do(func() {
		logger.Sync()
	})
}

func BenchmarkLoggingFactory_Logger_Parallel(b *testing.B) {
	once.Do(func() {
		//1. Create a new log server
		logging := NewLoggingFactory("debug", "flogging",
			WithModuleDate(), WithMaxFileSize(1024),
			WithMaxFileNum(5), WithRootDir("./tmp"))
		//2. Start writing files
		err := logging.Initial()
		if err != nil {
			b.Fatalf("init log faile, %v\n", err)
		}

		logger = logging.Logger("test")
	})

	//3. Log writing
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Debug("test debug ------------------")
		}
	})

	once.Do(func() {
		logger.Sync()
	})
}
