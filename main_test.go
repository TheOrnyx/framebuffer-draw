package main

import (
	"image"
	_ "image/png"
	"os"
	"testing"
	"time"

	"golang.org/x/sys/unix"
)

func BenchmarkBounce(b *testing.B) {
	x1 := 1080
	x2 := 1920
	x3 := 0
	x4 := 0
	x, y, w, h = &x4, &x3, &x2, &x1

	fd, err := os.OpenFile("/dev/fb0", os.O_RDWR, 0644)
	if err != nil {
		b.Fatalf("Failed to open framebuffer: %v", err)
	}
	defer fd.Close()

	mem, err := unix.Mmap(int(fd.Fd()), 0, *w**h*4, unix.PROT_WRITE, unix.MAP_SHARED)
	if err != nil {
		b.Fatalf("failed to mmap: %v", err)
	}

	file, err := os.Open("./sofa-cat.png")
	if err != nil {
		b.Fatalf("Unable to open image file: %v", err)
	}

	origMem = make([]byte, len(mem))
	copy(origMem, mem)
	defer file.Close()
	
	img, _, err := image.Decode(file)
	if err != nil {
		b.Fatalf("Unable to decode image: %v", err)
	}

	runFunc = runBounce
	
	startTime := time.Now()
	go runFunc(img, &mem, *x, *y)
	for time.Since(startTime) < time.Second*5 {
		
	}
}

func BenchmarkDraw(b *testing.B) {
	x1 := 1080
	x2 := 1920
	x3 := 0
	x4 := 0
	x, y, w, h = &x4, &x3, &x2, &x1

	fd, err := os.OpenFile("/dev/fb0", os.O_RDWR, 0644)
	if err != nil {
		b.Fatalf("Failed to open framebuffer: %v", err)
	}
	defer fd.Close()

	mem, err := unix.Mmap(int(fd.Fd()), 0, *w**h*4, unix.PROT_WRITE, unix.MAP_SHARED)
	if err != nil {
		b.Fatalf("failed to mmap: %v", err)
	}

	file, err := os.Open("./sofa-cat.png")
	if err != nil {
		b.Fatalf("Unable to open image file: %v", err)
	}

	origMem = make([]byte, len(mem))
	copy(origMem, mem)
	defer file.Close()
	
	img, _, err := image.Decode(file)
	if err != nil {
		b.Fatalf("Unable to decode image: %v", err)
	}

	runFunc(img, &mem, *x, *y)
}


