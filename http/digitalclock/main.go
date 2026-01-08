package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func main() {
	port := flag.Int("port", 8080, "Port to listen on")
	flag.Parse()

	http.HandleFunc("/", timeImageHandler)
	serverAddress := fmt.Sprintf(":%d", *port)
	fmt.Printf("Server listening on port %d...\n", *port)
	_ = http.ListenAndServe(serverAddress, nil)
}

func timeImageHandler(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	timeInput := queryParams.Get("time")
	scaleFactorInput := queryParams.Get("k")

	scaleFactor := 1
	if scaleFactorInput != "" {
		var err error
		scaleFactor, err = strconv.Atoi(scaleFactorInput)
		if err != nil || scaleFactor < 1 || scaleFactor > 30 {
			http.Error(w, "Invalid scale factor", http.StatusBadRequest)
			return
		}
	}

	if timeInput == "" {
		timeInput = time.Now().Format("15:04:05")
	} else {
		timeParts := strings.Split(timeInput, ":")
		if len(timeParts) != 3 {
			http.Error(w, "Invalid time format", http.StatusBadRequest)
			return
		}

		for _, part := range timeParts {
			if len(part) != 2 {
				http.Error(w, "Invalid time format", http.StatusBadRequest)
				return
			}
			if _, err := strconv.Atoi(part); err != nil {
				http.Error(w, "Invalid time format", http.StatusBadRequest)
				return
			}
		}

		hours, _ := strconv.Atoi(timeParts[0])
		minutes, _ := strconv.Atoi(timeParts[1])
		seconds, _ := strconv.Atoi(timeParts[2])

		if hours < 0 || hours > 23 {
			http.Error(w, "Invalid hours", http.StatusBadRequest)
			return
		}
		if minutes < 0 || minutes > 59 {
			http.Error(w, "Invalid minutes", http.StatusBadRequest)
			return
		}
		if seconds < 0 || seconds > 59 {
			http.Error(w, "Invalid seconds", http.StatusBadRequest)
			return
		}
	}

	image := createTimeImage(timeInput, scaleFactor)

	w.Header().Set("Content-Type", "image/png")
	png.Encode(w, image)
}

func createTimeImage(timeStr string, scaleFactor int) *image.RGBA {
	digits := []string{}
	for _, char := range timeStr {
		switch char {
		case '0':
			digits = append(digits, Zero)
		case '1':
			digits = append(digits, One)
		case '2':
			digits = append(digits, Two)
		case '3':
			digits = append(digits, Three)
		case '4':
			digits = append(digits, Four)
		case '5':
			digits = append(digits, Five)
		case '6':
			digits = append(digits, Six)
		case '7':
			digits = append(digits, Seven)
		case '8':
			digits = append(digits, Eight)
		case '9':
			digits = append(digits, Nine)
		case ':':
			digits = append(digits, Colon)
		}
	}

	digitHeight := strings.Count(Zero, "\n") + 1
	digitWidth := len(strings.Split(Zero, "\n")[0])
	colonWidth := len(strings.Split(Colon, "\n")[0])

	width := (6*digitWidth + 2*colonWidth) * scaleFactor
	height := digitHeight * scaleFactor

	img := image.NewRGBA(image.Rect(0, 0, width, height))

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.White)
		}
	}

	xPos := 0
	for _, digit := range digits {
		lines := strings.Split(digit, "\n")
		digitWidth := len(lines[0])

		for y, line := range lines {
			for x, char := range line {
				if char == '1' {
					for ky := 0; ky < scaleFactor; ky++ {
						for kx := 0; kx < scaleFactor; kx++ {
							img.Set(
								xPos+x*scaleFactor+kx,
								y*scaleFactor+ky,
								Cyan,
							)
						}
					}
				}
			}
		}

		xPos += digitWidth * scaleFactor
	}

	return img
}
