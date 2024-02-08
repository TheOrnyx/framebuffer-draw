package main

import (
	"flag"
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
	filePath string
)


func main() {
	w = flag.Int("width", 1920, "the max width to draw")
	h = flag.Int("height", 1080, "The max height to draw")
	flag.Parse()
	filePath = flag.Args()[0]

	fd, err := os.OpenFile("/dev/fb0", os.O_RDWR, 0644)
	if err != nil {
		log.Fatal("Failed to open framebuffer", err)
	}

	mem, err := unix.Mmap(int(fd.Fd()), 0, *w * *h * 4, unix.PROT_WRITE, unix.MAP_SHARED)
	if err != nil {
		log.Fatal("failed to mmap", err)
	}

	file, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("Unable to open image file: %v", err)
	}

	switch path.Ext(filePath) {
	case ".png", ".jpg", ".jpeg":
		img, _, err := image.Decode(file)
		if err != nil {
			log.Fatalf("Unable to decode image: %v", err)
		}

		drawPNG(img, &mem)

	case ".gif":
		img, err := gif.DecodeAll(file)
		if err != nil {
			log.Fatalf("Unable to decode image: %v", err)
		}

		drawGif(*img, &mem)
	}
	
	err = unix.Munmap(mem)
	if err != nil {
		log.Fatal("Failed to unmap mem,", err)
	}
}

// drawPNG draw a png to the framebuffer
func drawPNG(img image.Image, mem *[]byte)  {
	for y := 0; y < *h; y++ {
		if y > img.Bounds().Dy() {
			continue
		}
		for x := 0; x < *w; x++ {
			if x > img.Bounds().Dx() {
				continue
			}
			
			r, g, b, _ := img.At(x, y).RGBA()
			(*mem)[y * *w * 4 + x * 4 + 0] = byte(b)
			(*mem)[y * *w * 4 + x * 4 + 1] = byte(g)
			(*mem)[y * *w * 4 + x * 4 + 2] = byte(r)
		}
	}
}

// drawGif draw every frame of a gif image
func drawGif(gifImg gif.GIF, mem *[]byte)  {
	for  {
		for i := range gifImg.Image {
			img := gifImg.Image[i]
			for y := 0; y < *h; y++ {
				if y > img.Bounds().Dy() {
					continue
				}
				for x := 0; x < *w; x++ {
					if x > img.Bounds().Dx() {
						continue
					}
					
					r, g, b, _ := img.At(x, y).RGBA()
					(*mem)[y * *w * 4 + x * 4 + 0] = byte(b)
					(*mem)[y * *w * 4 + x * 4 + 1] = byte(g)
					(*mem)[y * *w * 4 + x * 4 + 2] = byte(r)
				}
			}
			time.Sleep((time.Millisecond * 10) * time.Duration(gifImg.Delay[i]))
			// time.Sleep(time.Millisecond * 250)
		}
	}
}
