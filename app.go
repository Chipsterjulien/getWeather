package main

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/anthonynsimon/bild/imgio"
	"github.com/anthonynsimon/bild/transform"
	"github.com/op/go-logging"
	"github.com/spf13/viper"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type config struct {
	Image []info
}

type info struct {
	Url     string
	PathSav string
	BaseUrl string
	Search  string
	Crop    string
	Resize  string
	Update  int
}

var C config

// voir http://stackoverflow.com/questions/38955605/reading-a-slice-of-maps-with-golang-viper

func main() {
	// confPath := "/etc/gincamalarm/"
	// confFilename := "gincamalarm"
	// logFilename := "/var/log/gincamalarm/error.log"

	confPath := "cfg/"
	confFilename := "getWeather"
	logFilename := "error.log"

	fd := initLogging(&logFilename)
	defer fd.Close()

	log := logging.MustGetLogger("log")

	loadConfig(&confPath, &confFilename)
	if err := viper.Unmarshal(&C); err != nil {
		log.Debug(err)
		os.Exit(1)
	}

	startApp()
}

func cropImage(crop *string, filePath *string) {
	log := logging.MustGetLogger("log")

	log.Debugf("Image will be crop to %s", *crop)

	img, err := imgio.Open(*filePath)
	if err != nil {
		log.Warningf("Unable to open image's file %s: %v", *filePath, err)
		return
	}

	cropTabStr := strings.Split(*crop, "x")
	if len(cropTabStr) != 4 {
		log.Warningf("In config file, \"%s\" is not good", cropTabStr)
		return
	}
	cropTabInt := make([]int, len(cropTabStr))
	for num, s := range cropTabStr {
		if cropTabInt[num], err = strconv.Atoi(s); err != nil {
			log.Warningf("Unable to convert \"%s\": %v", s, err)
			return
		}
	}

	ex := filepath.Ext(*filePath)
	log.Debugf("File extention is: %s", ex)
	result := transform.Crop(img, image.Rect(cropTabInt[0], cropTabInt[1], cropTabInt[2], cropTabInt[3]))

	var erro error
	switch ex {
	case ".png":
		log.Debug("png is detected")
		erro = imgio.Save(*filePath, result, imgio.PNG)
	case ".jpg":
		log.Debug("jpg is detected")
		erro = imgio.Save(*filePath, result, imgio.JPEG)
	}

	if erro != nil {
		log.Warningf("Unable to save image to %s: %v", *filePath, erro)
		return
	}

	log.Debug("Croping and saving is done")
}

func downloadAndSaveFile(url *string, filePath *string) error {
	log := logging.MustGetLogger("log")

	log.Debugf("Downloading %s ...", *url)
	out, err := os.Create(*filePath)
	if err != nil {
		return errors.New(fmt.Sprintf("Unable to create file \"%s\": %v", *filePath, err))
	}
	defer out.Close()

	resp, err := http.Get(*url)
	if err != nil {
		return errors.New(fmt.Sprintf("Unable to get url \"%s\": %v", *url, err))
	}
	defer resp.Body.Close()

	img, realExtension, err := image.Decode(resp.Body)
	if err != nil {
		return errors.New(fmt.Sprintf("Unable to decode image at \"%s\"", *url))
	}

	// nameOfFile := ""
	extensionGiven := filepath.Ext(*filePath)
	extensionGiven = strings.Replace(extensionGiven, ".", "", -1)

	log.Debugf("Extension's file: %s", realExtension)
	log.Debugf("Extension's given: %s", extensionGiven)

	if realExtension != extensionGiven {
		if extensionGiven != "jpg" {
			log.Warningf("File will be save to %s but the real extension is %s", *filePath, realExtension)
		}
	}

	switch extensionGiven {
	case "gif":
		f, err := os.Create(*filePath)
		if err != nil {
			return errors.New(fmt.Sprintf("Unable to create %s file: %v", *filePath, err))
		}
		defer f.Close()

		w := bufio.NewWriter(f)
		gif.Encode(w, img, &gif.Options{NumColors: 256})
		w.Flush()

	case "jpg":
		f, err := os.Create(*filePath)
		if err != nil {
			return errors.New(fmt.Sprintf("Unable to create %s file: %v", *filePath, err))
		}
		defer f.Close()

		w := bufio.NewWriter(f)
		jpeg.Encode(w, img, &jpeg.Options{Quality: 100})
		w.Flush()

	case "jpeg":
		f, err := os.Create(*filePath)
		if err != nil {
			return errors.New(fmt.Sprintf("Unable to create %s file: %v", *filePath, err))
		}
		defer f.Close()

		w := bufio.NewWriter(f)
		// jpeg.Encode(w, img, &jpeg.Options{Quality: jpeg.DefaultQuality})
		jpeg.Encode(w, img, &jpeg.Options{Quality: 100})
		w.Flush()

	case "png":
		f, err := os.Create(*filePath)
		if err != nil {
			return errors.New(fmt.Sprintf("Unable to create %s file: %v", *filePath, err))
		}
		defer f.Close()

		w := bufio.NewWriter(f)
		png.Encode(w, img)
		w.Flush()
	default:
		return errors.New(fmt.Sprintf("\"%s\" is an unknown format", extensionGiven))
	}

	log.Debug("Downloading and saving is done")

	return nil
}

func resizeImage(resize *string, filePath *string) {
	log := logging.MustGetLogger("log")

	log.Debugf("Image will be resize to %s", *resize)

	img, err := imgio.Open(*filePath)
	if err != nil {
		log.Warningf("Unable to open image's file %s: %v", *filePath, err)
		return
	}

	resizeTabStr := strings.Split(*resize, "x")
	if len(resizeTabStr) != 2 {
		log.Warningf("In config file, \"%s\" is not good", resizeTabStr)
		return
	}
	resizeTabInt := make([]int, len(resizeTabStr))
	for num, s := range resizeTabStr {
		if resizeTabInt[num], err = strconv.Atoi(s); err != nil {
			log.Warningf("Unable to convert \"%s\": %v", s, err)
			return
		}
	}

	ex := filepath.Ext(*filePath)
	// result := transform.Resize(img, resizeTabInt[0], resizeTabInt[1], transform.NearestNeighbor)
	result := transform.Resize(img, resizeTabInt[0], resizeTabInt[1], transform.Linear)
	var erro error

	switch ex {
	case ".png":
		log.Debug("png is detected")
		erro = imgio.Save(*filePath, result, imgio.PNG)
	case ".jpg":
		log.Debug("jpg is detected")
		erro = imgio.Save(*filePath, result, imgio.JPEG)
	}

	if erro != nil {
		log.Warningf("Unable to save image to %s: %v", *filePath, erro)
		return
	}

	log.Debug("Resizing is done")
}

func findRealUrl(val *info) error {
	resp, err := http.Get(val.Url)
	if err != nil {
		return errors.New(fmt.Sprintf("Unable to get url \"%s\": %v", val.Url, err))
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.New(fmt.Sprintf("Unable to read body for \"%s\": %v", err, val.Url))
	}

	for _, line := range strings.Split(string(body), "\n") {
		if strings.Contains(line, val.Search) {
			for _, word := range strings.Split(line, "\"") {
				if strings.Contains(word, val.Search) {
					if val.BaseUrl != "" {
						u, err := url.Parse(val.BaseUrl)
						if err != nil {
							return errors.New(fmt.Sprintf("Unable to parse \"%s\"", val.BaseUrl))
						}
						u.Path = filepath.Join(u.Path, word)
						val.Url = u.String()
					} else {
						val.Url = word
					}
					return nil
				}
			}
		}
	}

	return nil
}

func imageProcessing(numGo int, val info, done chan<- string) {
	log := logging.MustGetLogger("log")

	for true {
		if val.Search != "" {
			err := findRealUrl(&val)
			if err != nil {
				log.Warning(err)
				time.Sleep(2 * time.Minute)
				continue
			}
		}

		if err := downloadAndSaveFile(&val.Url, &val.PathSav); err != nil {
			log.Warning(err)

			time.Sleep(2 * time.Minute)
			continue
		}

		if val.Crop != "" {
			cropImage(&val.Crop, &val.PathSav)
		}

		if val.Resize != "" {
			resizeImage(&val.Resize, &val.PathSav)
		}
		log.Debugf("Goroutine %d. I'm waiting %ds before the next update", numGo, val.Update)
		time.Sleep(time.Duration(val.Update) * time.Second)
	}
	done <- fmt.Sprintf("function imageProcessing with \"%s\" url is terminated !", val.Url)
}

func startApp() {
	log := logging.MustGetLogger("log")
	done := make(chan string)

	for num, val := range C.Image {
		go imageProcessing(num+1, val, done)
	}

	log.Debugf("Program crashed since \"%s\" and have generate an unknown error", <-done)
}
