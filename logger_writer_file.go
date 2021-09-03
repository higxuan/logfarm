package logfarm

import (
	"fmt"
	"io/ioutil"
	"logfarm/files"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// MoveFileType move file type
type MoveFileType int

// MoveFileTypes
const (
	MoveFileTypeNothing MoveFileType = iota
	MoveFileTypePerMinute
	MoveFileTypeHourly
	MoveFileTypeDaily
)

type fileWriter struct {
	locker         sync.Mutex
	MaxLength      int64
	writeFile      string
	writeFileTime  time.Time
	FileName       string
	FileSuffix     string
	Separator      string
	TimeToMove     MoveFileType
	ticker         *time.Ticker
	lastMoveFlag   int
	stopChan       chan bool
	MaxBackupIndex int // 最大保留日志个数，如果为0则全部保留
}

var fileExecutor = files.New()

// NewFileWriter get file logger writer
func NewFileWriter(filename string, option *Option) LoggerWriter {
	fw := &fileWriter{
		FileName:  filename,
		Separator: "|",
		stopChan:  make(chan bool),
	}
	var err error
	if 0 == len(fw.FileName) {
		panic("filename not exist, input with param filename")
	}
	fw.writeFile = fw.FileName

	if option == nil {
		return fw
	}

	fw.MaxLength = option.FileMaxLength
	fw.FileSuffix = option.FileSuffix
	if 0 != len(fw.FileSuffix) {
		fw.writeFile += "." + fw.FileSuffix
	}

	_separator := option.Separator
	if 0 != len(_separator) {
		fw.Separator = _separator
	}

	_moveFileType := option.MoveFileType
	_maxBackupIndex := option.MaxBackupIndex

	fw.MaxBackupIndex = _maxBackupIndex

	fw.writeFileTime = time.Now()
	fi, err := fileExecutor.FileInfo(fw.writeFile)
	if err == nil {
		// 说明文件存在
		fw.writeFileTime = fi.ModTime()
	} else {
		// 没有文件创建文件
		fileExecutor.WriteAppend(fw.writeFile, "")
	}

	fw.TimeToMove = MoveFileType(_moveFileType)
	switch fw.TimeToMove {
	case MoveFileTypePerMinute:
		fw.lastMoveFlag = fw.writeFileTime.Minute()
	case MoveFileTypeHourly:
		fw.lastMoveFlag = fw.writeFileTime.Hour()
	case MoveFileTypeDaily:
		fw.lastMoveFlag = fw.writeFileTime.Day()
	}

	fw.ticker = time.NewTicker(time.Second)
	fw.timeToMoveFile()

	return fw
}

func (p *fileWriter) Write(values []string) (n int, err error) {

	p.locker.Lock()
	p.judgeMoveFile()
	p.locker.Unlock()

	n, err = fileExecutor.WriteAppend(p.writeFile, strings.Join(values, p.Separator)+"\n")
	if err != nil {
		return
	}

	if p.MaxLength == 0 {
		return
	}

	fi, e := fileExecutor.FileInfo(p.writeFile)
	if e != nil {
		return 0, e
	}

	if p.MaxLength > fi.Size() {
		return
	}

	p.moveFile(time.Now().Format("20060102T150405.999999999"))

	return
}

func (p *fileWriter) judgeMoveFile() error {
	timeStr, flag := "", 0
	timeNow := time.Now()
	switch p.TimeToMove {
	case MoveFileTypePerMinute:
		flag = timeNow.Minute()
		timeStr = p.writeFileTime.Format("200601021504")
	case MoveFileTypeHourly:
		flag = timeNow.Hour()
		timeStr = p.writeFileTime.Format("2006010215")
	case MoveFileTypeDaily:
		flag = time.Now().Day()
		timeStr = p.writeFileTime.Format("20060102")
	default:
		return nil
	}

	if flag == p.lastMoveFlag {
		return nil
	}
	p.lastMoveFlag = flag
	p.writeFileTime = time.Now()
	err := p.removeOldFiles()
	if err != nil {
		fmt.Println("remove old files error:", err.Error())
	}
	return p.moveFile(timeStr)
}

func (p *fileWriter) moveFile(timeStr string) error {
	filename := fmt.Sprintf("%s_%s", p.FileName, timeStr)
	if 0 != len(p.FileSuffix) {
		filename += "." + p.FileSuffix
	}
	err := fileExecutor.Rename(p.writeFile, filename)
	if err != nil {
		return err
	}

	_, err = fileExecutor.Write(p.writeFile, "")
	return err
}

func (p *fileWriter) timeToMoveFile() {
	go func() {
		for {
			select {
			case t := <-p.ticker.C:
				flag := 0
				switch p.TimeToMove {
				case MoveFileTypePerMinute:
					flag = t.Minute()
				case MoveFileTypeHourly:
					flag = t.Hour()
				case MoveFileTypeDaily:
					flag = t.Day()
				}
				if p.lastMoveFlag == flag {
					continue
				}
				p.locker.Lock()
				p.judgeMoveFile()
				p.locker.Unlock()
			case <-p.stopChan:
				return
			}
		}
	}()
}

func (p *fileWriter) removeOldFiles() error {
	if 0 == p.MaxBackupIndex {
		return nil
	}

	fileName := p.FileName

	if p.FileSuffix != "" {
		fileName += "." + p.FileSuffix
	}
	path := filepath.Dir(fileName)
	fileNameSplit := strings.Split(p.FileName, "/")
	filePrefix := fmt.Sprintf("%s_", fileNameSplit[len(fileNameSplit)-1])

	// 获取日志文件列表
	dirLis, err := ioutil.ReadDir(path)

	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	// 根据文件名过滤日志文件
	fileSort := FileSort{}
	for _, f := range dirLis {
		if strings.Contains(f.Name(), filePrefix) {
			fileSort = append(fileSort, f)
		}
	}

	if len(fileSort) <= p.MaxBackupIndex {
		return nil
	}

	// 根据文件修改日期排序，保留最近的N个文件
	sort.Sort(sort.Reverse(fileSort))
	for _, f := range fileSort[p.MaxBackupIndex:] {
		err := os.Remove(path + "/" + f.Name())
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *fileWriter) Stop() {
	p.stopChan <- true
}

// FileSort 文件排序
type FileSort []os.FileInfo

func (fs FileSort) Len() int {
	return len(fs)
}

func (fs FileSort) Less(i, j int) bool {
	return fs[i].Name() < fs[j].Name()
}

func (fs FileSort) Swap(i, j int) {
	fs[i], fs[j] = fs[j], fs[i]
}
