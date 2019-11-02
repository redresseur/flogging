package flogging

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
	"sync/atomic"
	"time"
)

const (
	DATE_DAY_FORMAT    = "2006-01-02"
	DATE_SECOND_FORMAT = "2006_01_02_15_04_05"
	log_suffix         = `.log`
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

	wt := &writer{
		//  the default of the signal capacity is 1
		data:         structure.NewQueue(1),
		WriterConfig: config,
	}

	wCtx, cancel := context.WithCancel(ctx)
	wt.ctx = context.WithValue(wCtx, wt, cancel)

	if wt.fbs, err = newFileBean(wt.Dir, wt.Prefix); err != nil {
		return
	}

	wt.write()
	return wt, nil
}

type writer struct {
	data *structure.Queue
	*WriterConfig
	fileDate time.Time
	fbs      *fileBean
	ctx      context.Context
}

func (w *writer) Write(p []byte) (n int, err error) {
	buffer := make([]byte, len(p), len(p)+1)
	copy(buffer, p)
	w.data.Push(buffer)
	w.data.SingleUP(false)
	return len(p), nil
}

func (w *writer) Sync() error {
	w.data.SingleUP(true)
	return nil
}

func (w *writer) Close() error {
	if v := w.ctx.Value(w); v != nil {
		if cancel, ok := v.(context.CancelFunc); ok {
			cancel()
		}
	}

	return nil
}

func (w *writer) flush() {
	for v := w.data.Pop(); v != nil; {
		if err := w.fileCheck(); err != nil {
			fmt.Printf("[file.Check] failure: %v", err)
			continue
		}

		n, err := w.fbs.Write(v.([]byte))
		if err != nil {
			fmt.Printf("[file.Write] failure: %v", err)
			continue
		}
		w.fbs.addSize(int64(n))
		v = w.data.Pop()
	}

	return
}

func (w *writer) write() {
	go func() {
		for {
			select {
			case <-w.ctx.Done():
				return
			case _, ok := <-w.data.Single():
				{
					if !ok {
						break
					}
					// write data into files
					w.flush()
					w.data.SingleDown()
				}
			}
		}
	}()
}

func (w *writer) statisticsLogFiles() (fs []string, err error) {
	var (
		fis []os.FileInfo
		re  *regexp.Regexp
	)

	if fis, err = ioutil.ReadDir(w.Dir); err != nil {
		return
	}

	expr := `^` + w.Prefix + `(.[a-zA-Z\_0-9]+)@(.[a-zA-Z\_0-9]+)` + log_suffix + `$`
	if re, err = regexp.Compile(expr); err != nil {
		return
	}

	for _, fi := range fis {
		if fi.IsDir() {
			continue
		}

		if subs := re.FindAllString(fi.Name(), -1); 0 == len(subs) {
			continue
		}

		fs = append(fs, fi.Name())
	}

	return
}

func (w *writer) fileCheck() error {
	if !w.isMustRename(w.fbs) {
		return nil
	}

	fs, err := w.statisticsLogFiles()
	if err != nil {
		return err
	}

	fb, err := newFileBean(w.Dir, w.Prefix)
	if err != nil {
		return err
	}

	// clean the files
	if nums := len(fs) + 1 - w.MaxFileCount; nums > 0 {
		for i := 0; i < nums; i++ {
			os.RemoveAll(path.Join(w.Dir, fs[i]))
		}
	}

	w.fbs = fb
	return nil
}

func (w *writer) isMustRename(fb *fileBean) bool {
	switch w.Model {
	case DateModel:
		if now().After(fb._date) {
			return true
		}
	case SizeModel:
		return fb.fileSize >= w.MaxSize
	}
	return false
}

type fileBean struct {
	path     string
	_date    time.Time
	fileSize int64
	*os.File
}

func newFileBean(dir, prefix string) (fb *fileBean, err error) {
	fb = &fileBean{}
	now := time.Now()
	fb.path = path.Join(dir, prefix) + now.Format(DATE_DAY_FORMAT) + log_suffix

	if fb._date, err = time.Parse(DATE_DAY_FORMAT, now.Format(DATE_DAY_FORMAT)); err != nil {
		return
	}

	fb.File, err = ioutils.OpenFile(fb.path, "")
	return
}

func (fb *fileBean) addSize(n int64) {
	atomic.AddInt64(&fb.fileSize, n)
}
