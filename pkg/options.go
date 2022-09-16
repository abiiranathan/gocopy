package gocopy

type option func(*GoCopier)

// Pass this option to constructor to turn on verbose mode
func Verbose() option {
	return func(wm *GoCopier) {
		wm.verbose = true
	}
}

func SkipIfExists() option {
	return func(w *GoCopier) {
		w.skipIfExists = true
	}
}

// Modify number of workers
func WithWorkers(n int) option {
	return func(w *GoCopier) {
		w.workers = n
	}
}

// modify the harsher function to uniquely idendify each file.
func WithCopier(copyFunc CopyFunc) option {
	return func(w *GoCopier) {
		w.copyFunction = copyFunc
	}
}
