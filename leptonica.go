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

// -------------- IMAGE FUNCTIONS -------------

func (p *goPix) PixFindSkew() (float32, float32) {
	var angle, conf C.l_float32
	C.pixFindSkew(p.cPix, &angle, &conf)
	return float32(angle), float32(conf)
}

func (p *goPix) PixFindSkewSlow() (float32, float32) {
	var angle, conf C.l_float32
	C.pixFindSkewSweepAndSearch(p.cPix, &angle, &conf, 1, 1, 10, 1, 0.01)
	return float32(angle), float32(conf)
}



