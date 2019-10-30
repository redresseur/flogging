package flogging

import (
	"fmt"
	structure "github.com/redresseur/utils/sturcture"
	"os"
	"sync/atomic"
	"time"
)

const DATE_FORMAT = "2006-01-02"

const (
	MODEL_DAY = iota
	MODEL_SIZE
)

type Writer struct {
	data *structure.Queue
}

func (w *Writer) Write(p []byte) (n int, err error) {
	buffer := make([]byte, len(p), len(p)+1)
	copy(buffer, p)
	w.data.Push(buffer)
	w.data.SingleUP(false)
	return len(p), nil
}

func (w *Writer) Sync() error {
	w.data.SingleUP(true)
	return nil
}

type WriteFile struct {
	data     *structure.Queue
	name string
	model    int
	fileDate time.Time
	fbs      *fileBean
}

func GetWriteFile(data *structure.Queue, name string) *WriteFile {
	return &WriteFile{data: data, name: name}
}

func (wf *WriteFile) StartSizeWrite(maxSize int64, maxFileCont int) {
	wf.model = MODEL_SIZE
	wf.fbs = newFileBean(wf.name, maxSize, maxFileCont)

	wf.write()
}

func (wf *WriteFile) StartDateWrite() {
	wf.model = MODEL_DAY
	wf.fbs = newFileBean(wf.name, 0, 0)

	wf.write()
}

func (wf *WriteFile) write() {
	go func() {
		for {
			select {
			case _, ok := <-wf.data.Single():
				{
					if !ok {
						break
					}

					for v := wf.data.Pop(); v != nil; {
						if err := wf.fileCheck(); err != nil {
							fmt.Print("err", string(v.([]byte)))
							continue
						}

						n, err := wf.fbs.write(v.([]byte))
						if err != nil {
							continue
						}
						wf.fbs.addSize(int64(n))

						v = wf.data.Pop()
					}

					wf.data.SingleDown()
				}
			}
		}
	}()
}

func (wf *WriteFile) fileCheck() error {
	if wf.isMustRename(wf.fbs) {
		return wf.fbs.rename(wf.model)
	}

	return nil
}

func (wf *WriteFile) isMustRename(fb *fileBean) bool {
	switch wf.model {
	case MODEL_DAY:
		t, _ := time.Parse(DATE_FORMAT, time.Now().Format(DATE_FORMAT))
		if t.After(*fb._date) {
			return true
		}
	case MODEL_SIZE:
		return fb.fileSize >= fb.maxFileSize
	}
	return false
}

type fileBean struct {
	name     string
	_date        *time.Time
	_suffix      int
	logFile      *os.File
	fileSize     int64
	maxFileSize  int64
	maxFileCount int
}

func newFileBean(name string, maxSize int64, maxFileCount int) (fb *fileBean) {
	t, _ := time.Parse(DATE_FORMAT, time.Now().Format(DATE_FORMAT))
	fb = &fileBean{name: name, _date: &t, maxFileSize: maxSize, maxFileCount: maxFileCount}
	fb.logFile, _ = os.OpenFile(name, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
	return
}

func (fb *fileBean) write(d []byte) (int, error) {
	return fb.logFile.Write(d)
}

func (fb *fileBean) addSize(n int64) {
	atomic.AddInt64(&fb.fileSize, n)
}

func (fb *fileBean) close() error {
	return fb.logFile.Close()
}

func (fb *fileBean) rename(model int) error {
	_ = fb.logFile.Close()
	var nextFileName string

	switch model {
	case MODEL_DAY:
		nextFileName = fmt.Sprint(fb.name, ".", fb._date.Format(DATE_FORMAT))
	case MODEL_SIZE:
		nextFileName = fmt.Sprint(fb.name, ".", fb.nextSuffix())
		fb._suffix = fb.nextSuffix()
	}

	if isExist(nextFileName) {
		if err := os.Remove(nextFileName); err != nil {
			return err
		}
	}
	err := os.Rename(fb.name, nextFileName)
	if err != nil {
		return err
	}

	t, _ := time.Parse(DATE_FORMAT, time.Now().Format(DATE_FORMAT))
	fb._date = &t
	if fb.logFile, err = os.OpenFile(fb.name, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666); err != nil {
		return err
	}
	fb.fileSize = 0
	return nil
}

func (fb *fileBean) nextSuffix() int {
	return int(fb._suffix%int(fb.maxFileCount) + 1)
}

func isExist(name string) bool {
	_, err := os.Stat(name)
	return err == nil || os.IsExist(err)
}

func fileSize(file string) int64 {
	f, e := os.Stat(file)
	if e != nil {
		fmt.Println(e.Error())
		return 0
	}
	return f.Size()
}
