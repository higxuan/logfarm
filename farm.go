package logfarm

import (
	"sync"
)

const (
	namespace = "Trellis::LogFarm"
)

// LogFarm functions to wite logs
type LogFarm interface {
	// WriteLog Write log into cache
	WriteLog(data []string) bool
	// Stop do not write data into file
	Stop()
}

type Option struct {
	FileName       string
	FileSuffix     string //log file' suffix
	ChanBuffer     int    //length of the log chan buffer
	FileMaxLength  int64  //the max length of log file, default: 0 is unlimited
	MoveFileType   int    // move file by per-minite(1) or hourly(2) or daily(3), 0 is doing nothing
	Separator      string
	MaxBackupIndex int
}

var mapLogger sync.Map

// New returns logFarm

func New(option *Option) LogFarm {
	fileName := option.FileName
	v, ok := mapLogger.Load(fileName)

	if ok && nil != v {
		return v.(LogFarm)
	}

	log := &logger{
		Writer: NewFileWriter(fileName, option),
	}
	log.logChan = make(chan []string, option.ChanBuffer)
	log.stopChan = make(chan bool)
	log.looperWriter()
	mapLogger.Store(fileName, log)
	return log
}
