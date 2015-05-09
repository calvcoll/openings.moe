package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"reflect"
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

var homeURL = "http://openings.moe/"
var directory = ""
var client = &http.Client{}
var wg sync.WaitGroup

func makeVideoAndSend(cs chan videoInfo) {
	req, _ := http.NewRequest("GET", (homeURL + "nextvideo.php"), nil)
	req.Header.Add("Host", "openings.moe")
	req.Header.Add("Referer", homeURL)
	resp, _ := client.Do(req)
	var body []byte
	if resp != nil {
		body, _ = ioutil.ReadAll(resp.Body)
	}
	// response := string(body)
	// fmt.Println(response)
	var video videoInfo
	err := json.Unmarshal(body, &video)
	if err != nil {
		fmt.Println("error:", err)
	}
	if _, err := os.Stat(directory + video.Videofname); os.IsNotExist(err) {
		cs <- video
	} else {
		fmt.Println(video.Videofname + " already exists, trying for another.")
		makeVideoAndSend(cs)
	}
}

func recieveVideoAndSave(cs chan videoInfo) {
	// needed because otherwise main will close before these are done.
	defer wg.Done()
	video := <-cs
	if video.Success {
		out, _ := os.Create(video.Videofname)
		defer out.Close()
		req, _ := http.NewRequest("GET", homeURL+video.Videourl, nil)
		req.Header.Add("Host", "openings.moe")
		req.Header.Add("Referer", homeURL)
		fileRequest, _ := client.Do(req)
		var file []byte
		if fileRequest != nil {
			file, _ = ioutil.ReadAll(fileRequest.Body)
		}
		out.Write(file)
		fmt.Printf("%+v has been successfully downloaded!\n", video.Videofname)
	}
}

func main() {
	args := os.Args[1:]

	j := 5

	if !reflect.DeepEqual(args, make([]string, 0)) {
		j64, _ := strconv.ParseInt(args[0], 10, 32) //sets j to be amount specified by args.
		j = int(j64)
	}

	cs := make(chan videoInfo, j) // becomes async, and doesn't block as badly.
	wg.Add(j)

	for i := 0; i < j; i++ {
		go makeVideoAndSend(cs)
		go recieveVideoAndSave(cs)
	}
	wg.Wait()
}
