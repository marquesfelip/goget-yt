package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"

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
	link := flag.String("l", "", "YouTube video URL")
	flag.Parse()

	if *link == "" {
		log.Fatalf(Red + "You must provide a YouTube video URL using the -l flag" + Reset)
	}

	client := youtube.Client{}

	video, err := client.GetVideo(*link)
	if err != nil {
		log.Fatalf(Red + "Failed to get video info: %v" + Reset, err)
	}

	formats := video.Formats
	for i, format := range formats {
		audio := "No"
		if format.AudioChannels > 0 {
			audio = "Yes"
		}
		fmt.Printf("[%d] Quality: %s, Audio: %s\n", i, format.QualityLabel, audio)
	}

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

	selectedFormat := formats[choice]

	stream, _, err := client.GetStream(video, &selectedFormat)
	if err != nil {
		log.Fatalf(Red + "Failed to get stream: %v" + Reset, err)
	}

	file, err := os.Create(video.Title + ".mp4")
	if err != nil {
		log.Fatalf(Red + "Failed to create file: %v" + Reset, err)
	}
	defer file.Close()

	fmt.Printf(Green+"Downloading: %v\n"+Reset, video.Title)
	buf := make([]byte, 32*1024)
	var downloaded int64
	if selectedFormat.ContentLength == 0 {
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
			if selectedFormat.ContentLength > 0 {
				fmt.Printf("\rDownloading... "+Blue+"%.2f%% complete"+Reset, float64(downloaded)/float64(selectedFormat.ContentLength)*100)
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
