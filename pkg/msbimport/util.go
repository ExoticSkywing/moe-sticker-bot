package msbimport

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

var httpClient = &http.Client{
	Timeout: 10 * time.Second,
}

func httpDownload(link string, f string) error {
	res, err := http.Get(link)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	fp, _ := os.Create(f)
	defer fp.Close()
	_, err = io.Copy(fp, res.Body)
	return err
}

func httpDownloadCurlUA(link string, f string) error {
	req, _ := http.NewRequest("GET", link, nil)
	req.Header.Set("User-Agent", "curl/7.61.1")
	res, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	fp, _ := os.Create(f)
	defer fp.Close()
	_, err = io.Copy(fp, res.Body)
	return err
}

func httpGet(link string) (string, error) {
	req, _ := http.NewRequest("GET", link, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.0.0 Safari/537.36")
	req.Header.Set("Accept-Language", "zh-Hant;q=0.9, ja;q=0.8, en;q=0.7")
	res, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	content, _ := io.ReadAll(res.Body)
	return string(content), nil
}

// redirected link, body, error
func httpGetWithRedirLink(link string) (string, string, error) {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", link, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.0.0 Safari/537.36")
	req.Header.Set("Accept-Language", "zh-Hant;q=0.9, ja;q=0.8, en;q=0.7")
	res, err := client.Do(req)
	if err != nil {
		return "", "", err
	}
	defer res.Body.Close()
	content, _ := io.ReadAll(res.Body)
	return res.Request.URL.String(), string(content), nil
}

func httpGetAndroidUA(link string) (string, error) {
	req, _ := http.NewRequest("GET", link, nil)
	req.Header.Set("User-Agent", "Android")
	res, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	content, _ := io.ReadAll(res.Body)
	return string(content), nil
}

func fDownload(link string, savePath string) error {
	cmd := exec.Command("curl", "-o", savePath, link)
	_, err := cmd.CombinedOutput()
	return err
}

func fExtract(f string) string {
	targetDir := filepath.Join(filepath.Dir(f), SecHex(4))
	os.MkdirAll(targetDir, 0755)
	log.Debugln("Extracting to :", targetDir)

	out, err := exec.Command(BSDTAR_BIN, "-xvf", f, "-C", targetDir).CombinedOutput()
	if err != nil {
		log.Errorln("Error extracting:", string(out))
		return ""
	} else {
		return targetDir
	}
}

func SecHex(n int) string {
	bytes := make([]byte, n)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func ArchiveExtract(f string) []string {
	targetDir := filepath.Join(path.Dir(f), SecHex(4))
	os.MkdirAll(targetDir, 0755)

	err := exec.Command(BSDTAR_BIN, "-xvf", f, "-C", targetDir).Run()
	if err != nil {
		log.Warnln("ArchiveExtract error:", err)
		return []string{}
	}
	return LsFilesR(targetDir, []string{}, []string{})
}

func LsFilesR(dir string, mustHave []string, mustNotHave []string) []string {
	var files []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		accept := true
		confidence := 0
		for _, kw := range mustHave {
			if !strings.Contains(strings.ToLower(path), strings.ToLower(kw)) {
				confidence += 1
			}
		}
		if confidence < len(mustHave) {
			accept = false
		}

		for _, kw := range mustNotHave {
			if strings.Contains(strings.ToLower(path), strings.ToLower(kw)) {
				accept = false
			}
		}
		if info.IsDir() {
			accept = false
		}
		log.Debugf("accept?: %t path: %s", accept, path)
		if accept {
			files = append(files, path)
		}
		return err
	})
	log.Debugln("listed following:")
	log.Debugln(files)
	if err != nil {
		return []string{}
	} else {
		return files
	}
}

func LsFiles(dir string, mustHave []string, mustNotHave []string) []string {
	var files []string
	glob, _ := filepath.Glob(path.Join(dir, "*"))

	for _, path := range glob {
		f, _ := os.Stat(path)
		if f.IsDir() {
			continue
		}

		accept := true
		for _, kw := range mustHave {
			if !strings.Contains(strings.ToLower(path), strings.ToLower(kw)) {
				accept = false
			}
		}
		for _, kw := range mustNotHave {
			if strings.Contains(strings.ToLower(path), strings.ToLower(kw)) {
				accept = false
			}
		}
		log.Debugf("accept?: %t path: %s", accept, path)
		if accept {
			files = append(files, path)
		}
	}
	return files
}

func FCompress(f string, flist []string) error {
	// strip data dir in zip.
	// comps are 2
	comps := "2"

	args := []string{"--strip-components", comps, "-avcf", f}
	// args := []string{"-avcf", f}
	args = append(args, flist...)

	log.Debugf("Compressing strip-comps:%s to file:%s for these files:%v", comps, f, flist)
	out, err := exec.Command(BSDTAR_BIN, args...).CombinedOutput()
	log.Debugln(string(out))
	if err != nil {
		log.Error("Compress error!")
		log.Errorln(string(out))
	}
	return err
}

func FCompressVol(f string, flist []string) []string {
	basename := filepath.Base(f)
	dir := filepath.Dir(f)
	zipIndex := 0
	var zips [][]string
	var zipPaths []string
	var curSize int64 = 0

	for _, f := range flist {
		st, err := os.Stat(f)
		if err != nil {
			continue
		}
		fSize := st.Size()
		if curSize == 0 {
			zips = append(zips, []string{})
		}
		if curSize+fSize < 50000000 {
			zips[zipIndex] = append(zips[zipIndex], f)
		} else {
			curSize = 0
			zips = append(zips, []string{})
			zipIndex += 1
			zips[zipIndex] = append(zips[zipIndex], f)
		}
		curSize += fSize
	}

	for i, files := range zips {
		var zipBN string
		if len(zips) == 1 {
			zipBN = basename
		} else {
			zipBN = strings.TrimSuffix(basename, ".zip")
			zipBN += fmt.Sprintf("_00%d.zip", i+1)
		}

		zipPath := filepath.Join(dir, zipBN)
		err := FCompress(zipPath, files)
		if err != nil {
			return nil
		}
		zipPaths = append(zipPaths, zipPath)
	}
	return zipPaths
}
