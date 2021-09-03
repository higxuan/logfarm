package files

import "os"

// FileRepo execute file functions
type FileRepo interface {
	FileOpened(string) bool                   // judge if file is opening
	Read(string) (b []byte, n int, err error) // read file
	Write(name, context string) (int, error)  // rewrite file with context
	WriteBytes(name string, b []byte) (int, error)
	WriteAppend(name, context string) (int, error) // append context to the file
	WriteAppendBytes(name string, b []byte) (int, error)
	Rename(oldPath, newPath string) error      // rename file
	SetReadBufLength(int) error                // set length of buffer to read file, default: 1024
	FileInfo(name string) (os.FileInfo, error) // get information with file name
}

// FileStatus defile file status
type FileStatus int

// file status
const (
	FileStatusClosed  FileStatus = iota // nothing
	FileStatusOpening                   // file is opened
	FileStatusMoving                    // file is moving or rename
)

// ReadBufferLength default reader buffer length
const (
	ReadBufferLength = 1024
)
