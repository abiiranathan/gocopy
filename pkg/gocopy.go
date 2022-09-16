package gocopy

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

// function for copying a file from src to destination dest.
type CopyFunc func(src string, dest string, overwrite bool) error

type GoCopier struct {
	workers int             // number of workers, default 4*runtime.GOMAXPROCS(0)
	limits  chan bool       // counting semaphore channel
	wg      *sync.WaitGroup // pointer because when wg is copied, it won't work.

	verbose      bool
	copyFunction CopyFunc // function for copying
	copyDest     string
	skipIfExists bool
}

func New(options ...option) *GoCopier {
	workers := runtime.GOMAXPROCS(0)

	wm := &GoCopier{
		workers:      workers,
		limits:       make(chan bool, workers),
		wg:           new(sync.WaitGroup),
		verbose:      false,
		skipIfExists: false,
	}

	if runtime.GOOS == "windows" {
		wm.copyFunction = copyWin32
	} else {
		wm.copyFunction = copyFile
	}

	for _, opt := range options {
		opt(wm)
	}

	return wm
}

func (wm *GoCopier) CopyDir(src, dst string) error {
	srcAbs, err := filepath.Abs(src)
	if err != nil {
		return err
	}

	dstAbs, err := filepath.Abs(dst)
	if err != nil {
		return err
	}

	wm.copyDest = dstAbs

	if err := os.MkdirAll(filepath.Join(dstAbs, filepath.Base(src)), 0666); err != nil {
		return err
	}

	go wm.searchTree(srcAbs)
	wm.wg.Add(1)
	wm.wg.Wait()
	return nil
}

// copy the file and when done, send its path on the pairs channel
func (wm *GoCopier) processFile(path string) {
	defer wm.wg.Done()

	// Wait on semaphore
	wm.limits <- true

	// Decrement counter when function completes
	defer func() {
		<-wm.limits
	}()

	dest := strings.Split(wm.copyDest, string(os.PathSeparator))
	splitPath := strings.Split(path, string(os.PathSeparator))

	// Normalize the destination path
	// filepath.Join does not replace similar prefix in path
	for i := 0; i < len(splitPath); i++ {
		if i == len(dest) {
			break
		}

		if dest[i] == splitPath[i] {
			continue
		}
		dest = append(dest, splitPath[i])
	}

	// join and clean the destination path
	abspath := filepath.Join(dest...)

	// Create tree of destination file if not exists
	if err := os.MkdirAll(filepath.Dir(abspath), 0666); err != nil {
		fmt.Printf("unable to create directory: %v\n", err)
		return
	}

	// Copy the file
	err := wm.copyFunction(path, abspath, !wm.skipIfExists)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Log the output after copy if verbose mode is on
	if wm.verbose {
		fmt.Println(path)
	}

}

func (wm *GoCopier) searchTree(dirname string) error {
	defer wm.wg.Done()

	// filepath.WalkDirFunc more performant than filepath.WalkFunc
	visitor := func(path string, d fs.DirEntry, err error) error {
		if err != nil && err != os.ErrNotExist {
			return err
		}

		fi, err := d.Info()
		if err != nil {
			return err
		}

		// ignore dir itself to avoid an infinite loop!
		if fi.Mode().IsDir() && path != dirname {
			wm.wg.Add(1)

			go wm.searchTree(path)
			return filepath.SkipDir
		}

		if fi.Mode().IsRegular() {
			wm.wg.Add(1)
			go wm.processFile(path)
		}

		return nil
	}

	// Wait on semaphore
	wm.limits <- true

	// Decrement semaphore counter when function exits
	defer func() {
		<-wm.limits
	}()

	return filepath.WalkDir(dirname, visitor)
}
