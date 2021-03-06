package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"sync"
)

type videoName struct {
	Title  string
	Source string
}

type videoInfo struct {
	Success    bool
	Videourl   string
	Videoname  videoName
	Videofname string
}

var domain = "openings.moe"
var homeURL = "http://" + domain + "/"
var directory = ""
var client = &http.Client{}
var wg sync.WaitGroup
var recursion = 0
var recursionLimit = 500

func makeVideoAndSend(cs chan videoInfo) {
	req, _ := http.NewRequest("GET", (homeURL + "nextvideo.php"), nil)
	req.Header.Add("Host", domain)
	req.Header.Add("Referer", homeURL)
	resp, _ := client.Do(req)
	var body []byte
	if resp != nil {
		body, _ = ioutil.ReadAll(resp.Body)
	}
	var video videoInfo
	err := json.Unmarshal(body, &video)
	if err != nil {
		fmt.Println("error:", err)
	}
	if _, err := os.Stat(directory + video.Videofname); os.IsNotExist(err) {
		cs <- video
	} else {
		recursion++
		if verbose {
			fmt.Println(video.Videofname + " already exists, trying for another.")
		}
		if recursion < recursionLimit {
			makeVideoAndSend(cs)
		} else {
			fakeVideo := videoInfo{false, "", videoName{"", ""}, ""}
			cs <- fakeVideo
		}
	}
}

func recieveVideoAndSave(cs chan videoInfo) {
	defer wg.Done()
	video := <-cs
	if video.Success {
		out, _ := os.Create(video.Videofname)
		defer out.Close()
		req, _ := http.NewRequest("GET", homeURL+video.Videourl, nil)
		req.Header.Add("Host", domain)
		req.Header.Add("Referer", homeURL)
		fileRequest, _ := client.Do(req)
		var file []byte
		if fileRequest != nil {
			file, _ = ioutil.ReadAll(fileRequest.Body)
		}
		out.Write(file)
		if !quiet {
			fmt.Printf("%+v has been successfully downloaded!\n", video.Videofname)
		}
	} else {
		fail++
	}
}

var verbose bool
var quiet bool
var buffer int
var fail = 0

var verboseFlag = flag.Bool("verbose", false, "Verboses outputs when file already downloaded.")
var quietFlag = flag.Bool("quiet", false, "Quietens output.")
var bufferFlag = flag.Int("buffer", 5, "Sets the amount of videos that can download simulaneously.")

func init() {
	flag.BoolVar(verboseFlag, "v", false, "Verboses outputs when file already downloaded.")
	flag.BoolVar(quietFlag, "q", false, "Quietens output.")
	flag.IntVar(bufferFlag, "b", 5, "Sets the amount of videos that can download simulaneously.")
}

func main() {
	flag.Parse()
	verbose = *verboseFlag
	quiet = *quietFlag
	buffer = *bufferFlag

	numStr := flag.Arg(0)
	if numStr == "" {
		numStr = "5" //so it doesn't throw an error when not specified
	}
	num, err := strconv.Atoi(numStr)

	if err == nil {
		cs := make(chan videoInfo, buffer) // becomes async, and doesn't block as badly.
		wg.Add(num)

		for i := 0; i < num; i++ {
			go makeVideoAndSend(cs)
			go recieveVideoAndSave(cs)
		}
		wg.Wait()
		if !quiet {
			if fail == 0 {
				fmt.Println("All " + strconv.Itoa(num) + " videos downloaded.")
			} else {
				fmt.Println(strconv.Itoa(fail) + " videos failed to download! Have you downloaded everything?")
			}
		}
	} else {
		fmt.Println("Select a number!")
	}
}
