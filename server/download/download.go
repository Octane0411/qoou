package download

import (
	"archive/zip"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func DownloadRepo(username, repoName string) {
	//https://github.com/Octane0411/qoou/archive/refs/heads/main.zip
	u := "https://github.com/" + username + "/" + repoName + "/archive/refs/heads/main.zip"

	//创建路径
	os.MkdirAll(GetDownloadsDir()+"/"+username, os.ModePerm)
	repoPath := filepath.Join(GetDownloadsDir(), username)
	//得到要下载文件的path
	downloadPath := filepath.Join(repoPath, "tmp.zip")
	//创建文件
	out, err := os.Create(downloadPath)
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()
	//访问url
	resp, err := http.Get(u)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	//下载文件
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	Unzip(downloadPath, repoPath)
	//TODO:这里可以优化
	//cp.Copy(repoPath+"/"+repoName+"-main", repoPath)
	//os.RemoveAll(repoPath + "/" + repoName + "-main")
	os.Rename(repoPath+"/"+repoName+"-main", repoPath+"/"+repoName)
	os.Remove(downloadPath)
}
func GetDownloadsDir() string {
	exe, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}
	dir := filepath.Dir(exe)
	return filepath.Join(dir, "downloads")
}
func GetRepoDir(username string, repoName string) string {
	return filepath.Join(GetDownloadsDir(), username, repoName)
}
func Unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer func() {
		if err := r.Close(); err != nil {
			panic(err)
		}
	}()

	os.MkdirAll(dest, 0755)

	// Closure to address file descriptors issue with all the deferred .Close() methods
	extractAndWriteFile := func(f *zip.File) error {
		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer func() {
			if err := rc.Close(); err != nil {
				panic(err)
			}
		}()

		path := filepath.Join(dest, f.Name)

		// Check for ZipSlip (Directory traversal)
		if !strings.HasPrefix(path, filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("illegal file path: %s", path)
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(path, f.Mode())
		} else {
			os.MkdirAll(filepath.Dir(path), f.Mode())
			f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}
			defer func() {
				if err := f.Close(); err != nil {
					panic(err)
				}
			}()

			_, err = io.Copy(f, rc)
			if err != nil {
				return err
			}
		}
		return nil
	}

	for _, f := range r.File {
		err := extractAndWriteFile(f)
		if err != nil {
			return err
		}
	}

	return nil
}
