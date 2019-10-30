package flogging

import (
	"testing"
)

func TestNewLoggingServer(t *testing.T) {
	//1. Create a new log server
	logging := NewLoggingServer("debug", "./test_new_log.txt", WithModuleDate(), WithMaxFileSize(1024), WithMaxFileNum(5))
	//2. Start writing files
	err := logging.StartWrite()
	if err != nil {
		t.Fatalf("init log faile, %v\n", err)
	}

	//3. Log writing
	for i:=0; i < 100; i++ {
		logging.Debug("test debug ------------------", i)
	}


	_ = logging.Sync()
}
