package gocopy

import (
	"io"
	"os"
	"syscall"
	"unsafe"
)

// Copies file at src to destination dest using io.Copy
// copyFile is cross-platform
func copyFile(src, dest string, overwrite bool) error {
	fsrc, err := os.Open(src)
	if err != nil {
		return err
	}
	defer fsrc.Close()

	fdst, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer fdst.Close()

	_, err = io.Copy(fdst, fsrc)
	return err
}

// Used if GOOS==windows
func copyWin32(src, dest string, overwrite bool) error {
	var (
		kernel32     = syscall.MustLoadDLL("kernel32.dll")
		copyFileProc = kernel32.MustFindProc("CopyFileW")
	)

	srcW, _ := syscall.UTF16FromString(src)
	dstW, _ := syscall.UTF16FromString(dest)

	var failIfExists uintptr = 1
	if overwrite {
		failIfExists = 0
	}

	rc, _, err := copyFileProc.Call(
		uintptr(unsafe.Pointer(&srcW[0])),
		uintptr(unsafe.Pointer(&dstW[0])),
		failIfExists,
	)

	if rc == 0 {
		return &os.PathError{
			Op:   "CopyFile",
			Path: src,
			Err:  err,
		}
	}
	return nil
}
