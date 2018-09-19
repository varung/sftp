package sftp

import (
	"io"
	"log"
	"os"
	"syscall"
	"time"
)

// DemoFS will support very limited operations
// It will allow listing and writing in an append only way
type DemoFileInfo struct {
}

func (d *DemoFileInfo) Name() string {
	return "bob"
}

func (d *DemoFileInfo) Size() int64 {
	return 10
}

func (d *DemoFileInfo) Mode() os.FileMode {
	return os.ModeDir
}

func (d *DemoFileInfo) IsDir() bool {
	return true
}

func (d *DemoFileInfo) ModTime() time.Time {
	return time.Now()
}

func (d *DemoFileInfo) Sys() interface{} {
	return nil
}

type DemoFile struct {
	f          *os.File
	dir        bool
	dir_cursor int
}

func (f *DemoFile) Stat() (os.FileInfo, error) {
	return &DemoFileInfo{}, nil
}
func (f *DemoFile) Close() error {
	return nil
}
func (f *DemoFile) ReadAt(b []byte, off int64) (n int, err error) {
	return 0, UnsupportedOpError{}
}
func (f *DemoFile) WriteAt(b []byte, off int64) (n int, err error) {
	return 0, UnsupportedOpError{}
}
func (f *DemoFile) Name() string {
	return "bob"
}
func (f *DemoFile) Chmod(mode os.FileMode) error {
	return UnsupportedOpError{}
}
func (f *DemoFile) Chown(uid, gid int) error {
	return UnsupportedOpError{}
}
func (f *DemoFile) Truncate(size int64) error {
	return UnsupportedOpError{}
}
func (f *DemoFile) Readdir(n int) ([]os.FileInfo, error) {
	log.Println(f.dir_cursor)
	if f.dir_cursor > 0 {
		return []os.FileInfo{}, io.EOF
	}
	f.dir_cursor++
	return []os.FileInfo{&DemoFileInfo{}}, nil
}

type DemoFS struct{}

func (v DemoFS) OpenDir(name string) (VFile, error) {
	log.Println("OpenDir", name)
	return &DemoFile{nil, true, 0}, nil
}
func (v DemoFS) OpenFile(name string, flag int) (VFile, error) {
	log.Println("OpenFile", name)
	return &DemoFile{nil, false, 0}, nil
}
func (v DemoFS) Stat(name string) (os.FileInfo, error) {
	log.Println("Stat", name)
	return &DemoFileInfo{}, nil
}
func (v DemoFS) Lstat(name string) (os.FileInfo, error) {
	log.Println("LStat", name)
	return &DemoFileInfo{}, nil
}
func (v DemoFS) Mkdir(name string, perm os.FileMode) error {
	log.Println("Mkdir", name)
	return UnsupportedOpError{}
}
func (v DemoFS) Remove(name string) error {
	log.Println("Remove", name)
	return UnsupportedOpError{}
}
func (v DemoFS) Rename(oldpath, newpath string) error {
	log.Println("Rename", oldpath, newpath)
	return UnsupportedOpError{}
}
func (v DemoFS) Symlink(oldname, newname string) error {
	return UnsupportedOpError{}
}
func (v DemoFS) Readlink(name string) (string, error) {
	return name, nil
}
func (v DemoFS) RealPath(path string) (string, error) {
	log.Println("Realpath", path, "=>", cleanPath(path))
	return cleanPath(path), nil
}
func (v DemoFS) Truncate(name string, size int64) error {
	log.Println("Truncate", name, size)
	return UnsupportedOpError{}
}
func (v DemoFS) Chmod(name string, mode uint32) error {
	log.Println("Chmod", name)
	return UnsupportedOpError{}
}
func (v DemoFS) Chtimes(name string, atime time.Time, mtime time.Time) error {
	log.Println("Chtimes", name)
	return UnsupportedOpError{}
}
func (v DemoFS) Chown(name string, uid, gid int) error {
	log.Println("Chown", name)
	return UnsupportedOpError{}
}

// creates SFTP error from DemoFS error
func (v DemoFS) StatusFromError(err error) StatusError {
	ret := StatusError{
		// ssh_FX_OK                = 0
		// ssh_FX_EOF               = 1
		// ssh_FX_NO_SUCH_FILE      = 2 ENOENT
		// ssh_FX_PERMISSION_DENIED = 3
		// ssh_FX_FAILURE           = 4
		// ssh_FX_BAD_MESSAGE       = 5
		// ssh_FX_NO_CONNECTION     = 6
		// ssh_FX_CONNECTION_LOST   = 7
		// ssh_FX_OP_UNSUPPORTED    = 8
		Code: ssh_FX_OK,
	}
	if err == nil {
		return ret
	}
	log.Printf("statusFromError: error is %T %#v", err, err)
	ret.Code = ssh_FX_FAILURE
	ret.msg = err.Error()

	switch e := err.(type) {
	case UnsupportedOpError:
		ret.Code = 8

	case syscall.Errno:
		ret.Code = translateErrno(e)
	case *os.PathError:
		log.Printf("statusFromError,pathError: error is %T %#v", e.Err, e.Err)
		if errno, ok := e.Err.(syscall.Errno); ok {
			ret.Code = translateErrno(errno)
		}
	case fxerr:
		ret.Code = uint32(e)
	default:
		switch e {
		case io.EOF:
			ret.Code = ssh_FX_EOF
		case os.ErrNotExist:
			ret.Code = ssh_FX_NO_SUCH_FILE
		}
	}
	return ret
}
