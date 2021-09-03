package logfarm_test

import (
	"logfarm"
	"testing"
	"time"
)

func TestLogFarm(t *testing.T)  {
	loggerTest := logfarm.New(&logfarm.Option{
		FileName:       "test",
		FileSuffix:     "log",
		ChanBuffer:     1000,
		MoveFileType:   2,
		Separator:      "",
		MaxBackupIndex: 7,
	})
	loggerTest.WriteLog([]string{"just test"})
	time.Sleep(3 * time.Second)
}
