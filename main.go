package main

import (
	"flag"
	"fmt"
	"image"
	"image/gif"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"os"
	"path"
	"time"

	"golang.org/x/sys/unix"
)

var (
	w *int
	h *int
	x *int
	y *int
	filePath string
	transThreshold uint32 = 0xF0F0 // the threshold for drawing transparent pixels (needs tweaking)
)

// initVars initialize the variables and return them
func initVars(imgPath string) (*[]byte, *os.File, error) {
	fd, err := os.OpenFile("/dev/fb0", os.O_RDWR, 0644)
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to open framebuffer: %v", err)
	}
	defer fd.Close()

	mem, err := unix.Mmap(int(fd.Fd()), 0, *w * *h * 4, unix.PROT_WRITE, unix.MAP_SHARED)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to mmap: %v", err)
	}

	file, err := os.Open(filePath)
	if err != nil {
		return nil, nil, fmt.Errorf("Unable to open image file: %v", err)
	}

	return &mem, file, nil
}

func main() {
	w = flag.Int("width", 1920, "the width of your framebuffer")
	h = flag.Int("height", 1080, "The height of your framebuffer")
	x = flag.Int("x", 0, "The start x position to draw at")
	y = flag.Int("y", 0, "The start y position to draw at")
	
	flag.Parse()
	filePath = flag.Args()[0] // TODO - make this like print out smth if there's no args

	mem, file, err := initVars(filePath)
	if err != nil {
		log.Fatalf("Failed to initialize program: %v", err)
	}
	defer file.Close()

	switch path.Ext(filePath) {
	case ".png", ".jpg", ".jpeg":
		img, _, err := image.Decode(file)
		if err != nil {
			log.Fatalf("Unable to decode image: %v", err)
		}

		drawPNG(img, mem)

	case ".gif":
		img, err := gif.DecodeAll(file)
		if err != nil {
			log.Fatalf("Unable to decode image: %v", err)
		}

		drawGif(*img, mem)
	}
	
	err = unix.Munmap(*mem)
	if err != nil {
		log.Fatal("Failed to unmap mem,", err)
	}
}

// drawPNG draw a png to the framebuffer
func drawPNG(img image.Image, mem *[]byte)  {
	drawImageAtPoint(img, mem, *x, *y)
}

// drawImageAtPoint draw the given image at a specific point
// x and y determine the top right position to draw from
func drawImageAtPoint(img image.Image, mem *[]byte, x, y int)  {
	for row := 0; row < img.Bounds().Dy(); row++ {
		if y + row >= *h {
			return
		}
		
		for col := 0; col < img.Bounds().Dx(); col++ {
			if x + col >= *w {
				break
			}
			
			r, g, b, a := img.At(col, row).RGBA()
			if a <= transThreshold {
				continue
			}
			
			(*mem)[(row + y) * *w * 4 + (col + x) * 4 + 0] = byte(b)
			(*mem)[(row + y) * *w * 4 + (col + x) * 4 + 1] = byte(g)
			(*mem)[(row + y) * *w * 4 + (col + x) * 4 + 2] = byte(r)
		}
	}
}

// drawGif draw every frame of a gif image
func drawGif(gifImg gif.GIF, mem *[]byte)  {
	for  {
		for i := range gifImg.Image {
			drawImageAtPoint(gifImg.Image[i], mem, 0, 0)
			fmt.Printf("")
			time.Sleep((time.Millisecond * 10) * time.Duration(gifImg.Delay[i]))
		}
	}
}
