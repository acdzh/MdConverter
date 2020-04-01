package main

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func main() {
	srcFilePath := os.Args[1]
	fileDir, err := filepath.Abs(filepath.Dir(srcFilePath))
	if err != nil {
		fmt.Println(err)
		return
	}

	fileStat, err := os.Stat(srcFilePath)
	if err != nil {
		fmt.Println(err)
		return
	} else if fileStat.IsDir() {
		fmt.Printf("Path is a directory.")
	}
	fileName := fileStat.Name()

	os.Chdir(fileDir)
	mdTxtByte, err := ioutil.ReadFile(fileName)
	if err != nil {
		fmt.Println("Read file failed.")
		return
	}
	mdTxt := string(mdTxtByte)

	result := regexp.MustCompile(`!\[(.*?)\]\((.*?)\)`).FindAllStringSubmatchIndex(mdTxt, -1)

	tags := "\n\n"
	for i := len(result) - 1; i >= 0; i-- {
		index := result[i]
		imgName := mdTxt[index[2]:index[3]]
		imgPath := mdTxt[index[4]:index[5]]
		imgTagName := fmt.Sprintf("%v-base64tag-%v", i, imgPath)
		fmt.Printf("Converting: %v\n", imgPath)
		imgBase64, err := imgPath2Base64(imgPath)
		if err != nil {
			fmt.Printf("failed: %v, %v, %v", i, mdTxt[index[0]:index[1]], err)
			continue
		}
		mdTxt = fmt.Sprintf("%s![%s][%s]%s", mdTxt[0:index[0]], imgName, imgTagName, mdTxt[index[1]:len(mdTxt)])
		tags += fmt.Sprintf("\n[%v]:data:image/png;base64,%v\n", imgTagName, imgBase64)
	}
	mdTxt += tags
	if ioutil.WriteFile(fmt.Sprintf("d-%v", fileName), []byte(mdTxt), 0644) == nil {
		fmt.Println("Success")
	}
}

func imgPath2Base64(imgPath string) (string, error) {
	stat, err := os.Stat(imgPath)
	if err == nil && !stat.IsDir() {
		return localImg2Base64(imgPath)
	} else if strings.HasPrefix(imgPath, "http") {
		return webImg2Base64(imgPath)
	}
	return "", fmt.Errorf("error")
}

func localImg2Base64(path string) (string, error) {
	imgFile, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer imgFile.Close()

	buf, err := ioutil.ReadAll(imgFile)
	if err != nil {
		return "", err
	}
	imgBase64str := base64.StdEncoding.EncodeToString(buf)
	return imgBase64str, nil
}

func webImg2Base64(url string) (string, error) {
	res, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	buf, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	imgBase64str := base64.StdEncoding.EncodeToString(buf)
	return imgBase64str, nil
}
