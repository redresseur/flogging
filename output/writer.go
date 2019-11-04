package output

import (
	"context"
	"fmt"
	"github.com/redresseur/utils/ioutils"
	"github.com/redresseur/utils/sturcture"
	"io"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"strconv"
	"sync/atomic"
	"time"
)

const (
	DATE_DAY_FORMAT    = "2006-01-02"
	DATE_SECOND_FORMAT = "2006_01_02_15_04_05"
	log_suffix         = `.log`

	DateModel = "date"
	SizeModel = "size"
)

func now() time.Time {
	t, _ := time.Parse(DATE_DAY_FORMAT, time.Now().Format(DATE_DAY_FORMAT))
	return t
}

type WriterConfig struct {
	Dir          string // path : the log directory
	Prefix       string // the log prefix
	Model        string // model: date or size
	MaxSize      int64  // maxSize: the max size of any file
	MaxFileCount int    // maxFileCount: the number of saved files
}

// Note: if the model is date, the maxSize and maxFileCount are not necessary.
func NewWriter(ctx context.Context, config *WriterConfig) (w io.Writer, err error) {
	if _, err = ioutils.CreateDirIfMissing(config.Dir); err != nil {
		return
	}

	fw := &fileWriter{
		//  the default of the signal capacity is 1
		data:         structure.NewQueue(1),
		WriterConfig: config,
	}

	wCtx, cancel := context.WithCancel(ctx)
	fw.ctx = context.WithValue(wCtx, fw, cancel)

	fw.statisticsLogFiles()
	fw.index++
	filePath := fmt.Sprintf("%s%s_%04d%s", path.Join(fw.Dir, fw.Prefix),
		now().Format(DATE_DAY_FORMAT), fw.index, log_suffix)

	if fw.fbs, err = newFileBean(filePath); err != nil {
		return
	}

	fw.write()
	return fw, nil
}

func Close(w io.Writer) error {
	if fw, ok := w.(*fileWriter); ok {
		if v := fw.ctx.Value(fw); v != nil {
			if cancel, ok := v.(context.CancelFunc); ok {
				cancel()
			}
		}
		if fw.fbs != nil && fw.fbs.File != nil {
			return fw.fbs.Close()
		}
	}

	return nil
}

type fileWriter struct {
	data *structure.Queue
	*WriterConfig
	fileDate time.Time
	fbs      *fileBean
	index    int
	ctx      context.Context
}

func (fw *fileWriter) Write(p []byte) (n int, err error) {
	buffer := make([]byte, len(p), len(p)+1)
	copy(buffer, p)
	fw.data.Push(buffer)
	fw.data.SingleUP(false)
	return len(p), nil
}

func (fw *fileWriter) Sync() error {
	fw.data.SingleUP(true)
	return nil
}

func (fw *fileWriter) Close() error {
	if v := fw.ctx.Value(fw); v != nil {
		if cancel, ok := v.(context.CancelFunc); ok {
			cancel()
		}
	}

	return nil
}

func (fw *fileWriter) flush() {
	for v := fw.data.Pop(); v != nil; {
		if err := fw.fileCheck(); err != nil {
			fmt.Printf("[file.Check] failure: %v", err)
			continue
		}

		n, err := fw.fbs.Write(v.([]byte))
		if err != nil {
			fmt.Printf("[file.Write] failure: %v", err)
			continue
		}
		fw.fbs.addSize(int64(n))
		v = fw.data.Pop()
	}

	return
}

func (fw *fileWriter) write() {
	go func() {
		for {
			select {
			case <-fw.ctx.Done():
				return
			case _, ok := <-fw.data.Single():
				{
					if !ok {
						break
					}
					// write data into files
					fw.flush()
					fw.data.SingleDown()
				}
			}
		}
	}()
}

func (fw *fileWriter) statisticsLogFiles() (fs []string, err error) {
	var (
		fis []os.FileInfo
		re  *regexp.Regexp
	)

	if fis, err = ioutil.ReadDir(fw.Dir); err != nil {
		return
	}

	expr := `^` + fw.Prefix + `([a-zA-Z0-9-]+)\_?([0-9]*)` + log_suffix + `$`
	if re, err = regexp.Compile(expr); err != nil {
		return
	}

	for _, fi := range fis {
		if fi.IsDir() {
			continue
		}

		if subs := re.FindStringSubmatch(fi.Name()); 0 == len(subs) {
			continue
		} else if len(subs) >= 3 {
			index, _ := strconv.Atoi(subs[2])
			if index > fw.index {
				fw.index = index
			}
		}

		fs = append(fs, fi.Name())
	}

	return
}

func (fw *fileWriter) fileCheck() error {
	if !fw.isMustRename(fw.fbs) {
		return nil
	}

	fs, err := fw.statisticsLogFiles()
	if err != nil {
		return err
	}

	newPath := fmt.Sprintf("%s%s_%04d%s", path.Join(fw.Dir, fw.Prefix),
		now().Format(DATE_DAY_FORMAT), fw.index, log_suffix)
	if fb, err := newFileBean(newPath); err != nil {
		return err
	} else {
		fw.fbs.Close()
		fw.fbs = fb
		fw.index++
	}

	// clean the files
	if nums := len(fs) + 1 - fw.MaxFileCount; nums > 0 {
		for i := 0; i < nums; i++ {
			os.RemoveAll(path.Join(fw.Dir, fs[i]))
		}
	}

	return nil
}

func (fw *fileWriter) isMustRename(fb *fileBean) bool {
	switch fw.Model {
	case DateModel:
		if now().After(fb._date) {
			return true
		}
	case SizeModel:
		return fb.fileSize >= fw.MaxSize
	}
	return false
}

type fileBean struct {
	path     string
	_date    time.Time
	fileSize int64
	*os.File
}

func newFileBean(path string) (fb *fileBean, err error) {
	fb = &fileBean{}
	now := time.Now()
	fb.path = path

	if fb._date, err = time.Parse(DATE_DAY_FORMAT, now.Format(DATE_DAY_FORMAT)); err != nil {
		return
	}

	fb.File, err = ioutils.OpenFile(fb.path, "")
	return
}

func (fb *fileBean) addSize(n int64) {
	atomic.AddInt64(&fb.fileSize, n)
}
