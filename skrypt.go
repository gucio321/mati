package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

const WorkingUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/94.0.4606.81 Safari/537.36"

var URL, DIR string

func main() {
	flag.StringVar(&URL, "url", "", "URL to download")
	flag.StringVar(&DIR, "dir", "", "Directory to save file")
	flag.Parse()
	if URL == "" || DIR == "" {
		flag.Usage()
		os.Exit(1)
	}

	client := &http.Client{}
	req, err := http.NewRequest("GET", URL, nil)
	if err != nil {
		log.Fatalf("unable to create request: %s", err)
	}

	req.Header.Set("User-Agent", WorkingUserAgent)
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("unable to do request: %s", err)
	}

	output := make([]byte, 0)
	// read all content of response.Body
	// into output
	for {
		buffer := make([]byte, 1024)
		n, err := resp.Body.Read(buffer)
		output = append(output, buffer[:n]...)
		if err != nil {
			break
		}
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(output)))
	if err != nil {
		log.Fatal(err)
	}

	// Find the div with the specified class
	div := doc.Find("div.typography")

	// Extract and print the content of the div
	fmt.Println(div.Text())

	imgURL, found := doc.Find("div.productFoto__main").Find("img").Attr("src")
	if !found {
		log.Fatal("Cannot extract imaage")
	}

	req, err = http.NewRequest("GET", imgURL, nil)
	if err != nil {
		log.Fatalf("unable to create request: %s", err)
	}

	req.Header.Set("User-Agent", WorkingUserAgent)
	resp, err = client.Do(req)
	if err != nil {
		log.Fatalf("unable to do request: %s", err)
	}

	if err = os.Mkdir(DIR, os.ModePerm); err != nil {
		log.Fatalf("unable to create directory: %s", err)
	}

	file, err := os.Create(filepath.Join(DIR, "image.jpg"))
	if err != nil {
		log.Fatalf("unable to create file: %s", err)
	}

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		log.Fatalf("unable to copy file: %s", err)
	}
}
