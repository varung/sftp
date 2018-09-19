package sftp

import (
	"os"
	"time"
)

// Defines interfaces that virtual filesystem should satisfy for sftp.Server

// A FileInfo describes a file and is returned by Stat and Lstat.
// type FileInfo interface {
// 	Name() string       // base name of the file
// 	Size() int65        // length in bytes for regular files; system-dependent for others
// 	Mode() FileMode     // file mode bits
// 	ModTime() time.Time // modification time
// 	IsDir() bool        // abbreviation for Mode().IsDir()
// 	Sys() interface{}   // underlying data source (can return nil)
// }

type VFile interface {
	Stat() (os.FileInfo, error)
	Close() error
	ReadAt(b []byte, off int64) (n int, err error)
	WriteAt(b []byte, off int64) (n int, err error)
	Name() string
	Chmod(mode os.FileMode) error
	Chown(uid, gid int) error
	Truncate(size int64) error

	// this is stateful: internally keep track of how many read so far
	Readdir(n int) ([]os.FileInfo, error)
}

type VFS interface {
	OpenDir(name string) (VFile, error)
	OpenFile(name string, flag int) (VFile, error)
	Stat(name string) (os.FileInfo, error)
	Lstat(name string) (os.FileInfo, error)
	Mkdir(name string, perm os.FileMode) error
	Remove(name string) error
	Rename(oldpath, newpath string) error
	Symlink(oldname, newname string) error
	Readlink(name string) (string, error)
	RealPath(path string) (string, error)
	Truncate(name string, size int64) error
	Chmod(name string, mode uint32) error
	Chtimes(name string, atime time.Time, mtime time.Time) error
	Chown(name string, uid, gid int) error
	// creates SFTP error from VFS error
	StatusFromError(err error) StatusError
}

type UnsupportedOpError struct{}

func (UnsupportedOpError) Error() string {
	return "Unsupported Operation"
}
