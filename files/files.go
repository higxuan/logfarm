package files

import (
	"io"
	"os"
	"sync"
)

// FileMode
const (
	FileModeOnlyRead  os.FileMode = 0444
	FileModeReadWrite os.FileMode = 0666
)

var defaultFile *fileGem

type fileGem struct {
	executingPath sync.Map
	readBufLength int
}

type callbackExec func() error

// New return fileRepo with default executor
func New() FileRepo {
	if defaultFile == nil {
		defaultFile = &fileGem{
			readBufLength: ReadBufferLength,
		}
	}
	return defaultFile
}

func (p *fileGem) updateExecFileStatus(name string, status FileStatus) error {
	if p.FileOpened(name) && status != FileStatusClosed {
		return ErrFileIsAlreadyOpen
	}
	if status == FileStatusClosed {
		p.executingPath.Delete(name)
		return nil
	}

	p.executingPath.Store(name, status)

	return nil
}

func (p *fileGem) Read(name string) (b []byte, n int, err error) {

	f := func() error {
		b, n, err = p.read(name, p.readBufLength)
		return err
	}

	err = p.execute(name, FileStatusOpening, f)

	return
}

func (p *fileGem) read(name string, bufLen int) (b []byte, n int, err error) {

	fi, e := p.tryOpen(name)
	if e != nil {
		err = e
		return
	}
	defer fi.Close()
	for {

		buf := make([]byte, bufLen)
		m, e := fi.Read(buf)
		if e != nil && e != io.EOF {
			err = ErrFailedReadFile
			return
		}
		n += m
		b = append(b, buf...)
		if m < bufLen {
			break
		}
	}

	return
}

func (p *fileGem) FileOpened(name string) bool {
	if v, ok := p.executingPath.Load(name); ok {
		return v.(FileStatus) != FileStatusClosed
	}
	return false
}

func (p *fileGem) tryOpen(name string) (*os.File, error) {
	return p.tryOpenFile(name, os.O_RDONLY, FileModeOnlyRead)
}

func (p *fileGem) tryOpenFile(name string, flag int, perm os.FileMode) (*os.File, error) {
	return os.OpenFile(name, flag, perm)
}

func (p *fileGem) Write(name, s string) (int, error) {
	return p.WriteBytes(name, []byte(s))
}

func (p *fileGem) WriteBytes(name string, b []byte) (int, error) {
	return p.write(name, b, os.O_TRUNC)
}

func (p *fileGem) WriteAppend(name, s string) (int, error) {
	return p.WriteAppendBytes(name, []byte(s))
}

func (p *fileGem) WriteAppendBytes(name string, b []byte) (int, error) {
	return p.write(name, b, os.O_APPEND)
}

func (p *fileGem) Rename(oldPath, newPath string) error {

	f := func() error {
		return os.Rename(oldPath, newPath)
	}

	return p.execute(oldPath, FileStatusMoving, f)
}

func (p *fileGem) SetReadBufLength(l int) error {
	if l <= 0 {
		return ErrReadBufferLengthBelowZero
	}

	p.readBufLength = l

	return nil
}

func (p *fileGem) write(name string, b []byte, flag int) (n int, err error) {

	callback := func() error {
		fi, e := p.tryOpenFile(name, os.O_CREATE|os.O_WRONLY|flag, FileModeReadWrite)
		if e != nil {
			return e
		}
		defer fi.Close()

		n, e = fi.Write(b)
		return e
	}

	err = p.execute(name, FileStatusOpening, callback)

	return
}

func (p *fileGem) execute(name string, fStatus FileStatus, callback callbackExec) (err error) {
	if err = p.updateExecFileStatus(name, fStatus); err != nil {
		return
	}
	defer p.updateExecFileStatus(name, FileStatusClosed)

	return callback()
}

func (p *fileGem) FileInfo(name string) (os.FileInfo, error) {
	return os.Stat(name)
}
