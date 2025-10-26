package main

import (
	"fmt"
	"image"
	_ "image/jpeg" // Register JPEG format
	_ "image/png"  // Register PNG format
	"os"
	"path/filepath"
)

func main() {
	// Check if an image file path was provided
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <image-file>\n", filepath.Base(os.Args[0]))
		os.Exit(1)
	}

	// Get the image file path from command line arguments
	imagePath := os.Args[1]

	// Open the image file
	file, err := os.Open(imagePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to open image file '%s': %v\n", imagePath, err)
		os.Exit(1)
	}
	defer file.Close()

	// Decode the image (format is auto-detected based on registered decoders)
	img, format, err := image.Decode(file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to decode image file '%s': %v\n", imagePath, err)
		fmt.Fprintf(os.Stderr, "Hint: Ensure the file is a valid PNG or JPEG image.\n")
		os.Exit(1)
	}

	// Get image dimensions
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Print the image dimensions
	fmt.Printf("Image format: %s\n", format)
	fmt.Printf("Width: %d pixels\n", width)
	fmt.Printf("Height: %d pixels\n", height)

	// Convert to ASCII art and print
	asciiArt := toASCII(img, 80)
	for _, line := range asciiArt {
		fmt.Println(line)
	}
}

// toASCII converts an image to ASCII art with the specified output width.
// The aspect ratio is preserved, accounting for typical terminal character height.
func toASCII(img image.Image, width int) []string {
	// ASCII palette from dark to light
	palette := "@%#*+=-:. "

	bounds := img.Bounds()
	imgWidth := bounds.Dx()
	imgHeight := bounds.Dy()

	// Calculate output height with character aspect ratio correction (~0.5)
	height := int(float64(imgHeight) * float64(width) / float64(imgWidth) * 0.5)

	// Prevent division by zero
	if height == 0 {
		height = 1
	}

	result := make([]string, height)

	// Process each row of the output ASCII art
	for y := 0; y < height; y++ {
		line := ""
		for x := 0; x < width; x++ {
			// Map ASCII coordinates back to image coordinates
			imgX := x * imgWidth / width
			imgY := y * imgHeight / height

			// Get pixel color and convert to grayscale
			r, g, b, _ := img.At(imgX, imgY).RGBA()
			// Use standard luminance formula (0.299*R + 0.587*G + 0.114*B)
			gray := (299*r + 587*g + 114*b) / 1000 / 256

			// Map brightness to ASCII character (invert: darker = dense chars)
			charIndex := int(gray) * (len(palette) - 1) / 255
			line += string(palette[charIndex])
		}
		result[y] = line
	}

	return result
}
