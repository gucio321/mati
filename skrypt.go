package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"golang.design/x/clipboard"
)

const WorkingUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/94.0.4606.81 Safari/537.36"

var (
	URL, DIR         string
	supportClipboard bool
	useClipboard     bool
)

func main() {
	// 0: initialization
	// 0.1: check if clipboard is supported
	// Init returns an error if the package is not ready for use.
	err := clipboard.Init()
	if err != nil {
		fmt.Println("Cannot use clipboard: ", err)
	} else {
		supportClipboard = true
	}

	// 0.2: load flags
	flag.StringVar(&URL, "url", "", "URL to download")
	flag.StringVar(&DIR, "dir", "", "Directory to save file")
	flag.BoolVar(&useClipboard, "c", false, "Use clipboard mechanism")
	flag.Parse()

	// 0.3: validate given URL/DIR; fallback prompt.
	if URL == "" { // fallback URL prompt
		fmt.Print("Enter URL: ")
		fmt.Scanln(&URL)
	}

	if DIR == "" && !useClipboard { // fallback DIR prompt
		fmt.Print("Enter DIR: ")
		fmt.Scanln(&DIR)
	}

	// 0.4: final validation of URL/DIR; display usage info if not successful.
	if URL == "" || (DIR == "" && !useClipboard) {
		flag.Usage()
		os.Exit(1)
	}

	// 1: fetching data
	// 1.1: create client and fetch HTML from the URL
	client := &http.Client{}
	req, err := http.NewRequest("GET", URL, nil)
	if err != nil {
		log.Fatalf("unable to create request: %s", err)
	}

	// this is important because you can get banned.
	req.Header.Set("User-Agent", WorkingUserAgent)
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("unable to do request: %s", err)
	}

	// 1.2: read response
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

	// 1.3: setup document analysis
	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(output)))
	if err != nil {
		log.Fatal(err)
	}

	// 1.4: extract description text
	// Find the div with the specified class
	div := doc.Find("div.typography")

	// Extract and print the content of the div
	divStr := strings.Split(div.Text(), "+48 660 887 000")[0] // remove contact info
	// 1.5: copy description to clipboard
	if supportClipboard && useClipboard {
		clipboard.Write(clipboard.FmtText, []byte(divStr))
		fmt.Print("Description copied, press enter to get table...")
		fmt.Scanln()
	} else {
		fmt.Println(divStr)
	}

	table := bytes.NewBufferString("-----")

	// Extract table
	div = doc.Find("div.productDetails__wrap").Find("div.productParams").Find("div.row")
	names := div.Find("div.productParams__name")
	params := div.Find("div.productParams__param")
	for i, d := range params.Nodes {
		fmt.Fprintf(table, "%s : %s\n", names.Nodes[i].FirstChild.Data, d.FirstChild.Data)
	}

	// 1.6: print table or paste to clipboard
	if supportClipboard && useClipboard {
		clipboard.Write(clipboard.FmtText, table.Bytes())
		fmt.Print("Table copied, press enter to download images...")
		fmt.Scanln()
	} else {
		fmt.Println(table.String())
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

			imgURLs = append(imgURLs, strings.ReplaceAll(a.Val, "foto_add_small", "foto_add_big"))
		}
	}

	if !useClipboard {
		if err = os.Mkdir(DIR, os.ModePerm); err != nil {
			log.Fatalf("unable to create directory: %s", err)
		}
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

		if supportClipboard && useClipboard {
			imgData := bytes.NewBufferString("")
			_, err = io.Copy(imgData, resp.Body)
			if err != nil {
				log.Fatalf("unable to copy image data: %s", err)
			}

			clipboard.Write(clipboard.FmtImage, imgData.Bytes())
			fmt.Printf("Image %d copied to clipboard, prese Enter for next...", i)
			fmt.Scanln()
		} else {
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
}
