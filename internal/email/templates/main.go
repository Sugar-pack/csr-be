package main

import (
	"encoding/base64"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/net/html"
)

func main() {
	err := processHTML("templates/registration-confirm/template.html", "templates/registration-confirm/index.html")
	if err != nil {
		panic(err)
	}
}

func processHTML(inputFile, outputFile string) error {
	// Load your HTML file
	htmlContent, err := os.ReadFile(inputFile)
	if err != nil {
		return err
	}

	// Parse the HTML content
	doc, err := html.Parse(strings.NewReader(string(htmlContent)))
	if err != nil {
		return err
	}

	// get path to the file to build proper image path
	directoryPath := filepath.Dir(inputFile)
	// Traverse and process HTML nodes
	err = traverseAndProcess(doc, directoryPath)
	if err != nil {
		return err
	}

	// Render the modified HTML and save it to output file
	var b strings.Builder
	html.Render(&b, doc)
	return os.WriteFile(outputFile, []byte(b.String()), 0644)
}

// traverseAndProcess traverses the HTML nodes and replaces image sources
func traverseAndProcess(n *html.Node, directoryFile string) error {
	if n.Type == html.ElementNode && n.Data == "img" {
		for i, a := range n.Attr {
			if a.Key == "src" && !strings.HasPrefix(a.Val, "http") {
				imgData, format, err := getImageData(a.Val, directoryFile)
				if err != nil {
					return err
				}
				n.Attr[i].Val = fmt.Sprintf("data:image/%s;base64,%s", format, imgData)
			}
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		err := traverseAndProcess(c, directoryFile)
		if err != nil {
			return err
		}
	}
	return nil
}

// getImageData fetches and converts an image to base64
func getImageData(url string, directoryFile string) (string, string, error) {
	var imageData []byte
	var err error

	// Image is a local file
	imageData, err = os.ReadFile(filepath.Join(directoryFile, url))

	if err != nil {
		return "", "", err
	}

	// Determine the image format
	format, err := getImageFormat(imageData)
	if err != nil {
		return "", "", err
	}

	// Encode to base64
	encoded := base64.StdEncoding.EncodeToString(imageData)
	return encoded, format, nil
}

// getImageFormat returns the format of the image
func getImageFormat(imageData []byte) (string, error) {
	// Simple check for SVG
	if isSVG(imageData) {
		return "svg+xml", nil
	}

	// Try decoding as a bitmap image
	_, format, err := image.DecodeConfig(strings.NewReader(string(imageData)))
	if err != nil {
		return "", err
	}

	return format, nil
}

// isSVG checks if the given data represents an SVG image
func isSVG(data []byte) bool {
	dataStr := string(data)
	return strings.Contains(dataStr, "<svg") || strings.HasSuffix(dataStr, ".svg")
}
