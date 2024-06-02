package fs

import (
	"path"
	"runtime"
)

const (
	dirConf       = "internal/config"
	dirMigrations = "db/migrations"
)

func Root() string {
	_, b, _, _ := runtime.Caller(0) //nolint:dogsled

	return path.Join(path.Dir(b), "/../../../")
}

// Config возвращает абсолютный путь до директории config.
func Config(filePath string) string {
	return makeDirPath(dirConf, filePath)
}

// Migrations возвращает абсолютный путь до директории migrations.
func Migrations(filePath string) string {
	return makeDirPath(dirMigrations, filePath)
}

func makeDirPath(dirPath string, filePath string) string {
	if filePath == "" {
		return path.Join(Root(), dirPath) + "/"
	}

	return path.Join(Root(), dirPath, filePath)
}
