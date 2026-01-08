//go:build !solution

package fileleak

import (
	"os"
	"path/filepath"
)

type testingT interface {
	Errorf(msg string, args ...interface{})
	Cleanup(func())
}

func VerifyNone(t testingT) {
	initialFiles := getOpenFiles(t)

	t.Cleanup(func() {
		checkForLeaks(t, initialFiles)
	})
}

// getOpenFiles возвращает множество открытых файлов
func getOpenFiles(t testingT) map[string]int {
	fds, err := os.ReadDir("/proc/self/fd")
	if err != nil {
		t.Errorf("failed to read /proc/self/fd: %v", err)
		return nil
	}

	files := make(map[string]int)
	for _, fd := range fds {
		linkPath := filepath.Join("/proc/self/fd", fd.Name())

		target, err := os.Readlink(linkPath)
		if err != nil {
			if !os.IsNotExist(err) {
				t.Errorf("failed to read link %s: %v", linkPath, err)
			}
			continue
		}
		files[target]++
	}

	return files
}

// checkForLeaks проверяет наличие новых файловых дескрипторов
func checkForLeaks(t testingT, initialFiles map[string]int) {
	currentFiles := getOpenFiles(t)
	if currentFiles == nil {
		return
	}

	for file, count := range currentFiles {
		if _, exists := initialFiles[file]; !exists || initialFiles[file] < count {
			t.Errorf("Leaked file: %s", file)
		}
	}
}
