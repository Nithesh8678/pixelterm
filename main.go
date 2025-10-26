package main

import (
	"flag"
	"fmt"
	"image"
	_ "image/jpeg" // Register JPEG format
	_ "image/png"  // Register PNG format
	"os"
	"strings"
)

func main() {
	// Define command-line flags
	width := flag.Int("width", 100, "output width in characters")
	scale := flag.Float64("scale", 0.15, "scale factor (affects height calculation)")
	color := flag.Bool("color", true, "enable colored ASCII output")
	save := flag.String("save", "", "save output to file instead of printing to stdout")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] <image-file>\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExample:\n")
		fmt.Fprintf(os.Stderr, "  %s -width 80 -color=false image.png\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -save output.txt image.jpg\n", os.Args[0])
	}

	flag.Parse()

	// Check if an image file path was provided
	if flag.NArg() < 1 {
		fmt.Fprintf(os.Stderr, "Error: No image file specified\n\n")
		flag.Usage()
		os.Exit(1)
	}

	imagePath := flag.Arg(0)

	// Open the image file
	file, err := os.Open(imagePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to open image file '%s': %v\n", imagePath, err)
		os.Exit(1)
	}
	defer file.Close()

	// Decode the image (format is auto-detected based on registered decoders)
	img, _, err := image.Decode(file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to decode image file '%s': %v\n", imagePath, err)
		fmt.Fprintf(os.Stderr, "Hint: Ensure the file is a valid PNG or JPEG image.\n")
		os.Exit(1)
	}

	// Generate ASCII art based on color flag
	var art []string
	if *color {
		art = colorASCII(img, *width, *scale)
	} else {
		art = toASCII(img, *width, *scale)
	}

	// Output to file or stdout
	if *save != "" {
		// Write to file
		output := strings.Join(art, "\n") + "\n"
		err := os.WriteFile(*save, []byte(output), 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Failed to write to file '%s': %v\n", *save, err)
			os.Exit(1)
		}
		fmt.Printf("ASCII art saved to '%s'\n", *save)
	} else {
		// Print to stdout
		for _, line := range art {
			fmt.Println(line)
		}
	}
}

// toASCII converts an image to ASCII art with the specified output width and scale.
// The aspect ratio is preserved, accounting for typical terminal character height.
func toASCII(img image.Image, width int, scale float64) []string {
	// ASCII palette from dark to light
	palette := "@%#*+=-:. "

	bounds := img.Bounds()
	imgWidth := bounds.Dx()
	imgHeight := bounds.Dy()

	// Calculate output height with character aspect ratio correction and scale
	height := int(float64(imgHeight) * float64(width) / float64(imgWidth) * scale)

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

// colorASCII converts an image to colored ASCII art using truecolor ANSI escapes.
// Character selection is based on grayscale, but colors are preserved from the original image.
func colorASCII(img image.Image, width int, scale float64) []string {
	// ASCII palette from dark to light
	palette := "@%#*+=-:. "

	bounds := img.Bounds()
	imgWidth := bounds.Dx()
	imgHeight := bounds.Dy()

	// Calculate output height with character aspect ratio correction and scale
	height := int(float64(imgHeight) * float64(width) / float64(imgWidth) * scale)

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

			// Get pixel color
			r, g, b, _ := img.At(imgX, imgY).RGBA()
			
			// Convert to 8-bit RGB values
			r8 := uint8(r >> 8)
			g8 := uint8(g >> 8)
			b8 := uint8(b >> 8)

			// Convert to grayscale for character selection
			gray := (299*r + 587*g + 114*b) / 1000 / 256

			// Map brightness to ASCII character
			charIndex := int(gray) * (len(palette) - 1) / 255
			char := palette[charIndex]

			// Build colored character with ANSI truecolor escape
			// Format: \x1b[38;2;<r>;<g>;<b>m<char>\x1b[0m
			line += fmt.Sprintf("\x1b[38;2;%d;%d;%dm%c\x1b[0m", r8, g8, b8, char)
		}
		result[y] = line
	}

	return result
}
