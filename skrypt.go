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
	fmt.Println("----")

	// Extract table
	div = doc.Find("div.productDetails__wrap").Find("div.productParams").Find("div.row")
	names := div.Find("div.productParams__name")
	params := div.Find("div.productParams__param")
	for i, d := range params.Nodes {
		fmt.Printf("%s : %s\n", names.Nodes[i].FirstChild.Data, d.FirstChild.Data)
	}

	// do image magic
	imgURLs := make([]string, 0)
	mainURL, found := doc.Find("div.productFoto__main").Find("img").Attr("src")
	if !found {
		log.Fatal("Cannot extract imaage")
	}
	imgURLs = append(imgURLs, mainURL)

	imgURL := doc.Find("div.productFoto__sliderList").Find("img") //.Attr("src")
	for _, u := range imgURL.Nodes {
		for _, a := range u.Attr {
			if a.Key != "src" {
				continue
			}
			imgURLs = append(imgURLs, a.Val)
		}
	}

	if err = os.Mkdir(DIR, os.ModePerm); err != nil {
		log.Fatalf("unable to create directory: %s", err)
	}

	for i, url := range imgURLs {
		req, err = http.NewRequest("GET", url, nil)
		if err != nil {
			log.Fatalf("unable to create request: %s", err)
		}

		req.Header.Set("User-Agent", WorkingUserAgent)
		resp, err = client.Do(req)
		if err != nil {
			log.Fatalf("unable to do request: %s", err)
		}

		file, err := os.Create(filepath.Join(DIR, fmt.Sprintf("image%d.jpg", i)))
		if err != nil {
			log.Fatalf("unable to create file: %s", err)
		}

		defer file.Close()

		_, err = io.Copy(file, resp.Body)
		if err != nil {
			log.Fatalf("unable to copy file: %s", err)
		}
	}
}
