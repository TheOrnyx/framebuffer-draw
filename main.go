package main

import (
	"flag"
	"fmt"
	"image"
	"image/draw"
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
	w              *int
	h              *int
	x              *int
	y              *int
	runType        *string // the type to run
	drawOver bool // whether the image should draw over the text
	filePath       string
	transThreshold uint8                                                  = 0xF0 // the threshold for drawing transparent pixels (needs tweaking)
	runFunc        func(img image.Image, mem *[]byte, startx, starty int) = drawImageAtPoint
	origMem        []byte
	imgRGBA *image.RGBA
)

// initVars initialize the variables and return them
func initVars(imgPath string) (*[]byte, *os.File, error) {
	fd, err := os.OpenFile("/dev/fb0", os.O_RDWR, 0644)
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to open framebuffer: %v", err)
	}
	defer fd.Close()

	mem, err := unix.Mmap(int(fd.Fd()), 0, *w**h*4, unix.PROT_WRITE, unix.MAP_SHARED)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to mmap: %v", err)
	}

	file, err := os.Open(filePath)
	if err != nil {
		return nil, nil, fmt.Errorf("Unable to open image file: %v", err)
	}

	origMem = make([]byte, len(mem))
	copy(origMem, mem)

	return &mem, file, nil
}

// chooseRunFunc choose and assign the run function based on the flag
func chooseRunFunc() {
	switch *runType {
	case "draw": // normal draw

	case "bounce": // bounce
		runFunc = runBounce
	default:
		log.Println("Unkown run function type, continuing with default draw")
	}
}

func main() {
	w = flag.Int("width", 1920, "the width of your framebuffer")
	h = flag.Int("height", 1080, "The height of your framebuffer")
	x = flag.Int("x", 0, "The start x position to draw at")
	y = flag.Int("y", 0, "The start y position to draw at")
	runType = flag.String("run", "draw", "the run type to draw\nOptions: draw, bounce")
	flag.BoolFunc("drawtop", "is present if the image should be drawn on top of text",
		func(s string) error {drawOver = true; return nil})

	flag.Parse()
	filePath = flag.Args()[0] // TODO - make this like print out smth if there's no arg
	chooseRunFunc()

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
		runFunc(img, mem, *x, *y)

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

// drawImageAtPoint draw the given image at a specific point
// x and y determine the top right position to draw from
func drawImageAtPoint(img image.Image, mem *[]byte, x, y int) {
	newMem := make([]byte, len(origMem))
	copy(newMem, origMem)
	bounds := img.Bounds()
	width, height := bounds.Max.X, bounds.Max.Y

	if imgRGBA == nil {
		makeImgRGBA(img)
	}
	
	for row := 0; row < width; row++ {
		if y+row >= *h {
			return
		}
		for col := 0; col < height; col++ {
			if x+col >= *w {
				break
			}

			index := (row*width + col) * 4
			pix := imgRGBA.Pix[index : index+4]
			if pix[3] <= transThreshold {
				continue
			}

			memIndex := (row+y)**w*4 + (col+x)*4
			if !drawOver && newMem[memIndex] != newMem[len(newMem)-4] {
				continue
			}

			newMem[memIndex] = pix[2]
			newMem[memIndex+1] = pix[1]
			newMem[memIndex+2] = pix[0]
		}
	}
	copy(*mem, newMem)
	// copy(unsafe.Slice((*byte)(unsafe.Pointer(&(*mem)[0])), len(newMem)), newMem)
	// unix.Msync(*mem, unix.MS_SYNC)
	fmt.Printf("") // doing a regular print like makes it like refresh faster for some reason?
}

// makeImgRGBA make an image rgba and assign the global object to it
func makeImgRGBA(img image.Image)  {
	bounds := img.Bounds()
	rgba := image.NewRGBA(bounds)
	draw.Draw(rgba, bounds, img, bounds.Min, draw.Src)
	imgRGBA = rgba
}

// drawGif draw every frame of a gif image
func drawGif(gifImg gif.GIF, mem *[]byte) {
	for {
		for i := range gifImg.Image {
			makeImgRGBA(gifImg.Image[i])
			drawImageAtPoint(gifImg.Image[i], mem, 0, 0)
			time.Sleep((time.Millisecond * 10) * time.Duration(gifImg.Delay[i]))
		}
	}
}
