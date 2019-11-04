package flogging

import (
	"sync"
	"testing"
)
var (
	once sync.Once
	testFactory *LoggingFactory
)

func TestNewLoggingFactory(t *testing.T) {
	//1. Create a new log server
	testFactory = NewLoggingFactory("debug", "flogging",
		WithModuleDate(), WithMaxFileSize(1024),
		WithMaxFileNum(5), WithRootDir("./tmp"))
	//2. Start writing files
	err := testFactory.Initial()
	if err != nil {
		t.Fatalf("init log faile, %v\n", err)
	}

	logger = testFactory.Logger("test")
	//3. Log writing
	for i := 0; i < 10240; i++ {
		logger.Debug("test debug ------------------", i)
	}

	logger.Sync()
}

func factoryTestInit()  {
	//1. Create a new log server
	testFactory = NewLoggingFactory("debug", "flogging",
		WithModuleSize(), WithMaxFileSize(1024),
		WithMaxFileNum(5), WithRootDir("./tmp"))
	//2. Start writing files
	err := testFactory.Initial()
	if err != nil {
		panic(err)
	}

	logger = testFactory.Logger("test")
}

func BenchmarkLoggingFactory_Logger(b *testing.B) {
	once.Do(factoryTestInit)
	//3. Log writing
	for i := 0; i < b.N; i++ {
		logger.Debug("test debug ------------------", i)
	}

	once.Do(func() {
		logger.Sync()
	})
}

func BenchmarkLoggingFactory_Logger_Parallel(b *testing.B) {
	once.Do(factoryTestInit)

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
