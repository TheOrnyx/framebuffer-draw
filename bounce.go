package main

import "image"

type imgBox struct {
	img  image.Image
	x    int
	y    int
	w    int
	h    int
	velx int
	vely int
}

// rightX return the right x position of the box
func (b *imgBox) rightX() int {
	return b.x + b.w
}

// btmY return the bottom y position of the box
func (b *imgBox) btmY() int {
	return b.y + b.h
}

// moveBox move the box by its velocity
func (b *imgBox) moveBox()  {
	// Update x position
	b.x += b.velx
	if b.x < 0 || b.rightX() > *w {
		b.velx = -b.velx
		if b.x < 0 {
			b.x = 0
		} else {
			b.x = *w - b.w
		}
	}

	// Update y position
	b.y += b.vely
	if b.y < 0 || b.btmY() > *h {
		b.vely = -b.vely
		if b.y < 0 {
			b.y = 0
		} else {
			b.y = *h - b.h
		}
	}
}

// drawBox draw the box to the screen
func (b *imgBox) drawBox(mem *[]byte)  {
	drawImageAtPoint(b.img, mem, b.x, b.y)
}

// makeBox create a new imgbox from an image and a startx and y
func makeBox(img image.Image, startx, starty int) *imgBox {
	return &imgBox{img: img,
		x:    startx,
		y:    starty,
		w:    img.Bounds().Dx(),
		h:    img.Bounds().Dy(),
		velx: 1,
		vely: 2}
}

// runBounce start running the bounce program
func runBounce(img image.Image, mem *[]byte, startx, starty int) {
	box := makeBox(img, startx, starty)
	for  {
		box.moveBox()
		box.drawBox(mem)
	}
}
