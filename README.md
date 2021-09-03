# logfarm
a tool to write logs to file with format data by separator
---

## Introduce

```go
// LogFarm functions to wite logs
type LogFarm interface {
	// Write log into cache
	WriteLog(data []string) bool
	// stop write data into file
	Stop()
}

logfarm.New(Option)
```

``` go
type Option struct {
	FileName       string
	FileSuffix     string //log file' suffix
	ChanBuffer     int    //length of the log chan buffer
	FileMaxLength  int64  //the max length of log file, default: 0 is unlimited
	MoveFileType   int    // move file by per-minite(1) or hourly(2) or daily(3), 0 is doing nothing
	Separator      string
	MaxBackupIndex int
}
```