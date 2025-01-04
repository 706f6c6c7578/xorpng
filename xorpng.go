package main

import (
	"crypto/rand"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"os"
	"path/filepath"
)

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [-i1 <first.png> -i2 <second.png>] [-g size -n count] > output.png\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "\nDescription:\n")
	fmt.Fprintf(os.Stderr, "  XORs two PNG images pixel by pixel or generates random noise images\n")
	fmt.Fprintf(os.Stderr, "\nFlags:\n")
	flag.PrintDefaults()
	fmt.Fprintf(os.Stderr, "\nExamples:\n")
	fmt.Fprintf(os.Stderr, "  XOR two images:\n")
	fmt.Fprintf(os.Stderr, "    %s -i1 image1.png -i2 image2.png > result.png\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "  Generate multiple random noise images:\n")
	fmt.Fprintf(os.Stderr, "    %s -g 480 -n 5\n", os.Args[0])
}

func imageToRGBA(src image.Image) *image.RGBA {
	bounds := src.Bounds()
	dst := image.NewRGBA(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := src.At(x, y).RGBA()
			dst.Set(x, y, color.RGBA{
				uint8(r >> 8),
				uint8(g >> 8),
				uint8(b >> 8),
				255,
			})
		}
	}
	return dst
}

func generateRandomImage(size int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	randomBytes := make([]byte, size*size*3)
	
	_, err := rand.Read(randomBytes)
	if err != nil {
		log.Fatalf("Error generating random data: %v", err)
	}

	idx := 0
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			r := randomBytes[idx]
			g := randomBytes[idx+1]
			b := randomBytes[idx+2]
			img.Set(x, y, color.RGBA{r, g, b, 255})
			idx += 3
		}
	}
	return img
}

func saveRandomImages(size, count int) {
	for i := 1; i <= count; i++ {
		filename := fmt.Sprintf("k-%d.png", i)
		f, err := os.Create(filename)
		if err != nil {
			log.Fatalf("Error creating file %s: %v", filename, err)
		}

		img := generateRandomImage(size)
		if err := png.Encode(f, img); err != nil {
			f.Close()
			log.Fatalf("Error encoding image %s: %v", filename, err)
		}
		
		f.Close()
		absPath, _ := filepath.Abs(filename)
		fmt.Printf("Generated: %s\n", absPath)
	}
}

func main() {
	img1Path := flag.String("i1", "", "Path to first PNG image")
	img2Path := flag.String("i2", "", "Path to second PNG image")
	genSize := flag.Int("g", 0, "Generate random noise image with specified size")
	numImages := flag.Int("n", 1, "Number of random images to generate")
	flag.Parse()

	if *genSize > 0 {
		saveRandomImages(*genSize, *numImages)
		return
	}

	if *img1Path == "" || *img2Path == "" {
		flag.Usage()
		os.Exit(1)
	}

	if fileInfo, _ := os.Stdout.Stat(); (fileInfo.Mode() & os.ModeCharDevice) != 0 {
		log.Fatal("Error: Output needs to be piped to a file")
	}

	img1File, err := os.Open(*img1Path)
	if err != nil {
		log.Fatalf("Error opening first image: %v", err)
	}
	defer img1File.Close()

	originalImg1, err := png.Decode(img1File)
	if err != nil {
		log.Fatalf("Error decoding first image: %v", err)
	}
	img1 := imageToRGBA(originalImg1)

	img2File, err := os.Open(*img2Path)
	if err != nil {
		log.Fatalf("Error opening second image: %v", err)
	}
	defer img2File.Close()

	originalImg2, err := png.Decode(img2File)
	if err != nil {
		log.Fatalf("Error decoding second image: %v", err)
	}
	img2 := imageToRGBA(originalImg2)

	bounds1 := img1.Bounds()
	bounds2 := img2.Bounds()
	if bounds1.Dx() != bounds2.Dx() || bounds1.Dy() != bounds2.Dy() {
		log.Fatalf("Error: Images have different dimensions\nImage1: %dx%d\nImage2: %dx%d",
			bounds1.Dx(), bounds1.Dy(), bounds2.Dx(), bounds2.Dy())
	}

	result := image.NewRGBA(bounds1)
	for i := 0; i < len(img1.Pix); i += 4 {
		result.Pix[i] = img1.Pix[i] ^ img2.Pix[i]         // R
		result.Pix[i+1] = img1.Pix[i+1] ^ img2.Pix[i+1]   // G
		result.Pix[i+2] = img1.Pix[i+2] ^ img2.Pix[i+2]   // B
		result.Pix[i+3] = 255                             // A
	}

	if err := png.Encode(os.Stdout, result); err != nil {
		log.Fatalf("Error encoding result image: %v", err)
	}
}
