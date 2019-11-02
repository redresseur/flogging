package flogging

import (
	"context"
	"fmt"
	"github.com/redresseur/utils/ioutils"
	"github.com/stretchr/testify/assert"
	"io"
	"os"
	"path"
	"sync"
	"testing"
)

var (
	testW         io.Writer
	testDir       = "./tmp"
	testPrefix    = "write_test"
	testModel     = SizeModel
	testSize      = int64(1024 * 1024)
	testFileCount = 10
	testLog       = path.Join(testDir, "test.log")
	once          sync.Once
)

func writeTestInit() {
	w, err := NewWriter(context.Background(), &WriterConfig{
		Dir:          testDir,
		Prefix:       testPrefix,
		Model:        testModel,
		MaxSize:      testSize,
		MaxFileCount: testFileCount,
	})

	if err != nil {
		panic(err)
	}

	testW = w
}

func TestWriter_Write(t *testing.T) {
	once.Do(writeTestInit)
	for i := 0; i < 1024; i++ {
		testW.Write([]byte(fmt.Sprintf("TestWriter_Write %v\n", i)))
	}

	testW.(*writer).Sync()
}

func TestWriter_Write_Date(t *testing.T) {
	testModel = DateModel
	once.Do(writeTestInit)
	for i := 0; i < 1024; i++ {
		testW.Write([]byte(fmt.Sprintf("TestWriter_Write %v\n", i)))
	}

	testW.(*writer).Sync()
}

func TestStatisticsLogFiles(t *testing.T) {
	once.Do(writeTestInit)
	testW.(*writer).statisticsLogFiles()
}

func BenchmarkWriter_Write(b *testing.B) {
	once.Do(writeTestInit)
	for i := 0; i < b.N; i++ {
		testW.Write([]byte(fmt.Sprintf("TestWriter_Write %v\n", i)))
	}

	once.Do(func() {
		testW.(*writer).Sync()
	})
}

func BenchmarkWriter_Write_Parallel(b *testing.B) {
	once.Do(writeTestInit)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			testW.Write([]byte(fmt.Sprintf("TestWriter_Write\n")))
		}
	})

	once.Do(func() {
		testW.(*writer).Sync()
	})
}

func BenchmarkWriter_File(b *testing.B) {
	fd, err := ioutils.OpenFile(testLog, "")
	assert.NoError(b, err)
	defer func() {
		fd.Close()
		os.RemoveAll(testLog)
	}()

	for i := 0; i < b.N; i++ {
		fd.Write([]byte(fmt.Sprintf("TestWriter_Write %v\n", i)))
	}
}

func BenchmarkWriter_File_Parallel(b *testing.B) {
	fd, err := ioutils.OpenFile(testLog, "")
	assert.NoError(b, err)
	defer func() {
		fd.Close()
		os.RemoveAll(testLog)
	}()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			fd.Write([]byte(fmt.Sprintf("TestWriter_Write\n")))
		}
	})
}
