package leptonica

/*
#cgo LDFLAGS: -llept
#include "leptonica/allheaders.h"
#include <stdlib.h>

l_uint8* uglycast(void* value) { return (l_uint8*)value; }

*/
import "C"
import (
	"errors"
	"sync"
	"unsafe"
)

type goPix struct {
	cPix   *C.PIX
	closed bool
	lock   sync.Mutex
}

// Deletes the pic, this must be called
func (p *goPix) Free() {
	p.lock.Lock()
	defer p.lock.Unlock()
	if !p.closed {
		// LEPT_DLL extern void pixDestroy ( PIX **ppix );
		C.pixDestroy(&p.cPix)
		C.free(unsafe.Pointer(p.cPix))
		p.closed = true
	}
}

// LEPT_DLL extern PIX * pixRead ( const char *filename );

// NewPixFromFile creates a new goPix from given filename
func NewPixFromFile(filename string) (*goPix, error) {
	cFilename := C.CString(filename)
	defer C.free(unsafe.Pointer(cFilename))

	// create new PIX
	CPIX := C.pixRead(cFilename)
	if CPIX == nil {
		return nil, errors.New("Unable to read file " + filename)
	}

	// all done
	pix := &goPix{
		cPix: CPIX,
	}
	return pix, nil
}

// NewPixReadMem creates a new goPix instance from a byte array
func NewPixReadMem(image *[]byte) (*goPix, error) {
	ptr := C.uglycast(unsafe.Pointer(&(*image)[0]))
	CPIX := C.pixReadMem(ptr, C.size_t(len(*image)))
	if CPIX == nil {
		return nil, errors.New("Not a valid image file")
	}
	pix := &goPix{
		cPix: CPIX,
	}
	return pix, nil
}

// NewPixReadMem creates a new goPix instance from a byte array
func (p *goPix) PixWriteMemPnm() (data, error) {
	
	var b C.l_uint8
	var size C.size_t
	err := C.pixWriteMemPnm(&b, &size, p.cPix)
	data := []uint8(b)
	C.free(unsafe.Pointer(b))
	C.free(unsafe.Pointer(size))
	if err == 1 {
		return data, errors.New(`Failed writing PBM to bytes`)
	}
	return data, nil
	
}

// -------------- IMAGE FUNCTIONS -------------

func (p *goPix) SkewAngle() (float32, float32) {
	var angle, conf C.l_float32
	C.pixFindSkew(p.cPix, &angle, &conf)
	a, c := float32(angle), float32(conf)
	C.free(unsafe.Pointer(angle))
	C.free(unsafe.Pointer(conf))
	return a, c
}

func (p *goPix) SkewAngleSlow() (float32, float32) {
	var angle, conf C.l_float32
	C.pixFindSkewSweepAndSearch(p.cPix, &angle, &conf, 1, 1, 10, 1, 0.01)
	a, c := float32(angle), float32(conf)
	C.free(unsafe.Pointer(angle))
	C.free(unsafe.Pointer(conf))
	return a, c
}

func (p *goPix) OrientationAngle() (*goPix, float32, int, error) {
	var a, c C.l_float32
	newpix := C.pixDeskewGeneral(p.cPix, 1, 7, 0.01, 1, 0, &a, &c)
	if newpix == nil {
		return p, 0, 0, errors.New(`Deskew failed`)
	}
	p.Free()
	var upconf, leftconf C.l_float32
	err := C.pixOrientDetect(newpix, &upconf, &leftconf, 0, 0)
	if err == 1 {
		C.pixDestroy(&newpix)
		C.free(unsafe.Pointer(newpix))
		return nil, 0, 0, errors.New(`Orientation detection failed`)
	}
	var orient C.l_int32
	err = C.makeOrientDecision(upconf, leftconf, 0.0, 0.0, &orient, 0)
	if err == 1 {
		C.pixDestroy(&newpix)
		C.free(unsafe.Pointer(newpix))
		return nil, 0, 0, errors.New(`Orientation decision failed`)
	}
	
	radians := float32(a)
	orientation := int(orient)
	switch orientation {
		case 2: radians += 1.57079633 // left-facing
				tmp := C.pixRotate90(newpix, 1)
				if tmp != newpix && tmp != nil {
					C.pixDestroy(&newpix)
					C.free(unsafe.Pointer(newpix))
					newpix = tmp
				}
		case 3: radians += 3.14159265 // upside-down
				tmp := C.pixRotate180(newpix, newpix)
				if tmp != newpix && tmp != nil {
					C.pixDestroy(&newpix)
					C.free(unsafe.Pointer(newpix))
					newpix = tmp
				}
		case 4: radians += 4.71238898 // right-facing
				tmp := C.pixRotate90(newpix, -1)
				if tmp != newpix && tmp != nil {
					C.pixDestroy(&newpix)
					C.free(unsafe.Pointer(newpix))
					newpix = tmp
				}
	}
	for radians > 6.28318531 {
		radians -=  6.28318531
	}
	
	pix := &goPix{
		cPix: newpix,
	}
	
	return pix, radians, orientation, nil
}

