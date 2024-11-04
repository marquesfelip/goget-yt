package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/kkdai/youtube/v2"
)

const (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
)

func main() {
	link := parseFlags()
	client := youtube.Client{}

	video := getVideoInfo(client, link)
	formats := video.Formats

	// Filter formats to only include those with quality information
	var filteredFormats youtube.FormatList
	for _, format := range formats {
		if format.QualityLabel != "" {
			filteredFormats = append(filteredFormats, format)
		}
	}

	// Sort formats by quality, prioritizing those with audio
	sort.Slice(filteredFormats, func(i, j int) bool {
		resI := parseQualityLabel(filteredFormats[i].QualityLabel)
		resJ := parseQualityLabel(filteredFormats[j].QualityLabel)

		if filteredFormats[i].AudioChannels != filteredFormats[j].AudioChannels {
			return filteredFormats[i].AudioChannels > filteredFormats[j].AudioChannels
		}
		return resI > resJ
	})

	displayFormats(filteredFormats)
	choice := getUserChoice(filteredFormats)

	selectedFormat := filteredFormats[choice]
	stream := getStream(client, video, selectedFormat)

	downloadVideo(stream, video.Title, selectedFormat)
}

// parseFlags parses the command line flags and returns the YouTube video URL.
func parseFlags() string {
	link := flag.String("l", "", "YouTube video URL")
	flag.Parse()

	if *link == "" {
		log.Fatalf(Red + "You must provide a YouTube video URL using the -l flag" + Reset)
	}
	return *link
}

// getVideoInfo retrieves the video information from YouTube.
func getVideoInfo(client youtube.Client, link string) *youtube.Video {
	video, err := client.GetVideo(link)
	if err != nil {
		log.Fatalf(Red + "Failed to get video info: %v" + Reset, err)
	}
	return video
}

// displayFormats displays the available formats for the video.
func displayFormats(formats youtube.FormatList) {
	for i, format := range formats {
		audio := "No"
		if format.AudioChannels > 0 {
			audio = "Yes"
			fmt.Printf(Green+"[%d] Quality: %s, Audio: %s"+Reset+"\n", i, format.QualityLabel, audio)
		} else {
			fmt.Printf("[%d] Quality: %s, Audio: %s\n", i, format.QualityLabel, audio)
		}
	}
}

// getUserChoice prompts the user to select a format and returns the chosen index.
func getUserChoice(formats youtube.FormatList) int {
	fmt.Print("Enter the number of the format you want to download: ")
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		log.Fatalf(Red + "Failed to read input: %v" + Reset, err)
	}

	choice, err := strconv.Atoi(input[:len(input)-1])
	if err != nil || choice < 0 || choice >= len(formats) {
		log.Fatalf(Red + "Invalid choice" + Reset)
	}
	return choice
}

// getStream retrieves the video stream for the selected format.
func getStream(client youtube.Client, video *youtube.Video, format youtube.Format) io.ReadCloser {
	stream, _, err := client.GetStream(video, &format)
	if err != nil {
		log.Fatalf(Red + "Failed to get stream: %v" + Reset, err)
	}
	return stream
}

// downloadVideo downloads the video stream to a file.
func downloadVideo(stream io.ReadCloser, title string, format youtube.Format) {
	file, err := os.Create(title + ".mp4")
	if err != nil {
		log.Fatalf(Red + "Failed to create file: %v" + Reset, err)
	}
	defer file.Close()

	fmt.Printf(Green+"Downloading: %v\n"+Reset, title)
	buf := make([]byte, 32*1024)
	var downloaded int64
	if format.ContentLength == 0 {
		fmt.Println(Yellow + "Unknown video size, progress will be shown in megabytes." + Reset)
	}
	for {
		n, err := stream.Read(buf)
		if n > 0 {
			_, err := file.Write(buf[:n])
			if err != nil {
				log.Fatalf(Red + "Failed to write to file: %v" + Reset, err)
			}
			downloaded += int64(n)
			if format.ContentLength > 0 {
				fmt.Printf("\rDownloading... "+Blue+"%.2f%% complete"+Reset, float64(downloaded)/float64(format.ContentLength)*100)
			} else {
				fmt.Printf("\rDownloading... "+Blue+"%.2f MB complete"+Reset, float64(downloaded)/(1024*1024))
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf(Red + "Failed to download video: %v" + Reset, err)
		}
	}

	fmt.Println(Green + "\nDownload completed!" + Reset)
}

// parseQualityLabel extracts the numeric resolution from QualityLabel.
func parseQualityLabel(label string) int {
	resStr := strings.TrimSuffix(label, "p")
	resInt, err := strconv.Atoi(resStr)
	if err != nil {
		return 0
	}
	return resInt
}
