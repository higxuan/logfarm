package logfarm

// LoggerWriter logger writer repo
type LoggerWriter interface {
	Write([]string) (int, error)
	Stop()
}
