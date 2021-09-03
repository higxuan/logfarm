package logfarm

// logger implements for logger writer
type logger struct {
	Writer   LoggerWriter
	logChan  chan []string
	stopChan chan bool
}

// WriteLog write logs to the filename
func (p *logger) WriteLog(data []string) bool {
	p.logChan <- data
	return true
}

// Stop do not write data into file
func (p *logger) Stop() {
	p.stopChan <- true
}

func (p *logger) looperWriter() {
	go func() {
		for {
			select {
			case log := <-p.logChan:
				p.Writer.Write(log)
			case <-p.stopChan:
				p.Writer.Stop()
				close(p.logChan)
				close(p.stopChan)
				return
			}
		}
	}()
}
