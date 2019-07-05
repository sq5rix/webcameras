package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func main() {

	lst := getList("cameras.lst")

	outputDirName := "Views"
	os.Mkdir(outputDirName, 0755)

	for _, v := range lst {
		if v != "" {
			fmt.Println(v)
			if isPix(v) {
				openPix(v, outputDirName)
			} else {
				openMap(v, outputDirName)
			}
		}
	}

	pollInterval := 15

	timerCh := time.Tick(time.Duration(pollInterval) * time.Minute)

	for range timerCh {

		lst := getList("cameras.lst")
		for _, v := range lst {
			if v != "" {
				fmt.Println(v)
				if isPix(v) {
					openPix(v, outputDirName)
				} else {
					openMap(v, outputDirName)
				}
			}

		}
	}
}
func openPix(pix, DirName string) {
	nm := getExp(pix, `/[_a-z0-9A-Z]+.jpg`)
	DownloadFile(filepath.Join(DirName, setStampedName(nm)), pix)
}

func openMap(site, DirName string) {
	agent := "Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:67.0) Gecko/20100101 Firefox/67.0"
	body, err := HTTPRequestCustomUserAgent(site, agent)
	if err != nil {
		log.Fatal(err)
	}
	ht := getExp(string(body), `poster=[:"/._a-z0-9A-Z"]+.style`)
	if ht == "" {
		ht = getExp(string(body), `poster=[:"/._a-z0-9A-Z"]+.preload=`)
	}
	ht = getExp(ht, `https[:"/._a-z0-9A-Z"]+jpg`)
	nm := getExp(ht, `/[_a-z0-9A-Z]+.jpg`)

	DownloadFile(filepath.Join(DirName, setStampedName(nm)), ht)
}

func getTag(pageContent, startTag, endTag string) string {

	titleStartIndex := strings.Index(pageContent, startTag)
	if titleStartIndex == -1 {
		fmt.Printf("No %s element found\n", startTag)
		os.Exit(0)
	}
	titleStartIndex += len(startTag)

	// Find the index of the closing tag
	titleEndIndex := strings.Index(pageContent, endTag)
	if titleEndIndex == -1 {
		fmt.Printf("No %s found.\n", endTag)
		os.Exit(0)
	}
	return pageContent[titleStartIndex:titleEndIndex]
}

// Create a regular expression to find comments
func getExp(pageContent, expr string) string {
	re := regexp.MustCompile(expr)
	sl := re.FindString(pageContent)
	return sl
}

// HTTPRequestCustomUserAgent with agent string
func HTTPRequestCustomUserAgent(url, userAgent string) (b []byte, err error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		err = errors.New(
			"resp.StatusCode: " +
				strconv.Itoa(resp.StatusCode))
		return
	}

	return ioutil.ReadAll(resp.Body)
}

// DownloadFile will download a url to a local file. It's efficient because it will
// write as it downloads and not load the whole file into memory.
func DownloadFile(filepath string, url string) error {

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}

func getList(lst string) []string {

	lines := make([]string, 1)

	file, err := os.Open(lst)
	if err != nil {
		log.Fatal("Provide cameras.lst file")
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		ln := scanner.Text()
		lines = append(lines, ln)
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	return lines
}

func isPix(v string) bool {
	return v[len(v)-3:] == "jpg"
}

func setStampedName(nm string) string {
	t := time.Now()
	tm := t.Format(time.RFC3339)
	return nm[1:len(nm)-4] + "_" + tm[:len(tm)-6] + ".jpg"
}
