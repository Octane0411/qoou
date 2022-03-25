package download

import (
	"log"
	"os"
	"path/filepath"
)

func DownloadRepo(username, repoName string) {
	//https://github.com/Octane0411/qoou/archive/refs/heads/main.zip

}

func GetDownloadsDir() string {
	exe, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}
	dir := filepath.Dir(exe)
	return filepath.Join(dir, "downloads")
}
