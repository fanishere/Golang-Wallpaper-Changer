package main

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

// RedditResponse is Reddit outer response data
type RedditResponse struct {
	MetaData struct {
		Modhash string     `json:"modhash"`
		Dist    int        `json:"dist"`
		Posts   []PostData `json:"children"`
	} `json:"data"`
}

// PostData is Reddit inner data for individual posts
type PostData struct {
	Post struct {
		Title  string `json:"title"`
		Link   string `json:"url"`
		Domain string `json:"domain"`
	} `json:"data"`
}

func main() {
	setWallpaper(getRandomImage(getRedditPosts()))
}

// getRedditPosts gets the top 25 posts from /r/wallpaper,
// and parses them into structs
func getRedditPosts() ([]PostData, error) {
	url := "https://www.reddit.com/r/wallpaper/hot/.json?"
	client := http.Client{}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "for-reddit-wallpaper-changer-komorrr")

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	data := RedditResponse{}
	if err = json.NewDecoder(res.Body).Decode(&data); err != nil {
		return nil, err
	}

	return data.MetaData.Posts, nil
}

// getRandomImage gets a random image from the list of reddit posts
// and calls downloadImage on it to download
func getRandomImage(posts []PostData, err error) (string, error) {
	if err != nil {
		return "", err
	}

	rand.Seed(time.Now().Unix())
	var imagePost *PostData

	for imagePost == nil {
		randPost := posts[rand.Intn(len(posts))]
		if (randPost.Post.Domain == "i.imgur.com") || (randPost.Post.Domain == "i.redd.it") {
			imagePost = &randPost
		}

	}

	return downloadImage(imagePost.Post.Link, imagePost.Post.Title)

}

func downloadImage(imgURL, imgTitle string) (string, error) {
	filename := ImageFileName(imgTitle)
	currentDir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	os.Chdir(filepath.Join(currentDir, "wallpapers"))
	out, err := os.Create(filename)
	if err != nil {
		return "", err
	}
	defer out.Close()

	res, err := http.Get(imgURL)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	_, err = io.Copy(out, res.Body)
	if err != nil {
		return "", err
	}

	return filename, nil
}

// ImageFileName takes a Reddit Post Title and turns it into an image filename like "wallpaper.jpg"
func ImageFileName(imgTitle string) string {

	windowsIllegal := regexp.MustCompile("[<>:\"/\\|?,*]")
	cleanedImageTitle := windowsIllegal.ReplaceAllString(imgTitle, "")

	filename := strings.Replace(cleanedImageTitle, " ", "-", -1) + ".jpg"

	return filename
}

func setWallpaper(filename string, err error) error {
	if err != nil {
		fmt.Println(err)
		return err
	}

	cacheDir, _ := os.Getwd()
	filepath := filepath.Join(cacheDir, filename)

	user32 := windows.NewLazyDLL("user32.dll")
	systemParametersInfo := user32.NewProc("SystemParametersInfoW")
	filenameUTF16, err := windows.UTF16PtrFromString(filepath)

	if err != nil {
		return err
	}
	systemParametersInfo.Call(
		uintptr(0x0014),                        //uiAction = pointer to set desktop wallpaper
		uintptr(0x0000),                        //uiparam = 0
		uintptr(unsafe.Pointer(filenameUTF16)), //pointer to wallpaper file
		uintptr(0x01|0x02),                     //fWinIni broadcasts change to user profile spiUpdateINIFile | spifSendChange
	)

	return nil
}
