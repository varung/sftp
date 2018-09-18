package sftp

import (
	"io"
	"os"
	"path/filepath"
	"syscall"
	"time"
)

type VFile struct {
	f *os.File
}
type VFileInfo = os.FileInfo

func (f VFile) Stat() (VFileInfo, error) {
	return f.f.Stat()
}
func (f VFile) Close() error {
	return f.f.Close()
}
func (f VFile) ReadAt(b []byte, off int64) (n int, err error) {
	return f.f.ReadAt(b, off)
}
func (f VFile) WriteAt(b []byte, off int64) (n int, err error) {
	return f.f.WriteAt(b, off)
}
func (f VFile) Name() string {
	return f.f.Name()
}
func (f VFile) Chmod(mode os.FileMode) error {
	return f.f.Chmod(mode)
}
func (f VFile) Chown(uid, gid int) error {
	return f.f.Chown(uid, gid)
}
func (f VFile) Truncate(size int64) error {
	return f.f.Truncate(size)
}
func (f VFile) Readdir(n int) ([]VFileInfo, error) {
	return f.f.Readdir(n)
}

type VFS struct{}

var vfs VFS

func (v VFS) OpenDir(name string) (VFile, error) {
	f, err := os.OpenFile(name, os.O_RDONLY, 0644)
	return VFile{f}, err
}
func (v VFS) OpenFile(name string, flag int) (VFile, error) {
	f, err := os.OpenFile(name, flag, 0644)
	return VFile{f}, err
}
func (v VFS) Stat(name string) (os.FileInfo, error) {
	return os.Stat(name)
}
func (v VFS) Lstat(name string) (os.FileInfo, error) {
	return os.Lstat(name)
}
func (v VFS) Mkdir(name string, perm os.FileMode) error {
	return os.Mkdir(name, perm)
}
func (v VFS) Remove(name string) error {
	return os.Remove(name)
}
func (v VFS) Rename(oldpath, newpath string) error {
	return os.Rename(oldpath, newpath)
}
func (v VFS) Symlink(oldname, newname string) error {
	return os.Symlink(oldname, newname)
}
func (v VFS) Readlink(name string) (string, error) {
	return os.Readlink(name)
}
func (v VFS) RealPath(path string) (string, error) {
	f, err := filepath.Abs(path)
	f = cleanPath(f)
	return f, err
}
func (v VFS) Truncate(name string, size int64) error {
	return os.Truncate(name, size)
}
func (v VFS) Chmod(name string, mode uint32) error {
	return os.Chmod(name, os.FileMode(mode))
}
func (v VFS) Chtimes(name string, atime time.Time, mtime time.Time) error {
	return os.Chtimes(name, atime, mtime)
}
func (v VFS) Chown(name string, uid, gid int) error {
	return os.Chown(name, uid, gid)
}

// translateErrno translates a syscall error number to a SFTP error code.
func translateErrno(errno syscall.Errno) uint32 {
	switch errno {
	case 0:
		return ssh_FX_OK
	case syscall.ENOENT:
		return ssh_FX_NO_SUCH_FILE
	case syscall.EPERM:
		return ssh_FX_PERMISSION_DENIED
	}
	return ssh_FX_FAILURE
}

type UnsupportedOpError struct{}

func (UnsupportedOpError) Error() string {
	return "Unsupported Operation"
}

// creates SFTP error from VFS error
func (v VFS) StatusFromError(err error) StatusError {
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
	debug("statusFromError: error is %T %#v", err, err)
	ret.Code = ssh_FX_FAILURE
	ret.msg = err.Error()

	switch e := err.(type) {
	case UnsupportedOpError:
		ret.Code = 8

	case syscall.Errno:
		ret.Code = translateErrno(e)
	case *os.PathError:
		debug("statusFromError,pathError: error is %T %#v", e.Err, e.Err)
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
