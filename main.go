package main

import (
	"flag"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"os"

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
	
	img, err := openImage(filePath)
	if err != nil {
		log.Fatal("Failed to open image:", err)
	}
	

	for y := 0; y < *h; y++ {
		if y > img.Bounds().Dy() {
			continue
		}
		for x := 0; x < *w; x++ {
			if x > img.Bounds().Dx() {
				continue
			}
			
			r, g, b, _ := img.At(x, y).RGBA()
			mem[y * *w * 4 + x * 4 + 0] = byte(b)
			mem[y * *w * 4 + x * 4 + 1] = byte(g)
			mem[y * *w * 4 + x * 4 + 2] = byte(r)
		}
	}

	err = unix.Munmap(mem)
	if err != nil {
		log.Fatal("Failed to unmap mem,", err)
	}
}

// openImage open, decode and return and image based on the path
func openImage(path string) (image.Image, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("Unable to open image file: %v", err)
	}

	img, _, err := image.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("Unable to decode image: %v", err)
	}

	return img, nil
}
