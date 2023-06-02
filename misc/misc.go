package misc

import (
	"math"
	"os"
	"path/filepath"
	"strconv"
)

// if limit is not found in env, we set to max
func EnvToLimit(key string) (int64, error) {
	val := os.Getenv(key)
	var intVal int64 = math.MaxInt64
	if val != "" {
		var err error
		intVal, err = strconv.ParseInt(val, 10, 64)
		if err != nil {
			return 0, err
		}
	}
	return intVal, nil
}

func DiskUsage(path string) (int64, error) {
	var totalBytes int64
	err := filepath.Walk(path, func(fpath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		totalBytes += info.Size()
		return nil
	})
	if err != nil {
		return 0, err
	}
	return totalBytes, nil
}
