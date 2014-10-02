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

type pixStruct struct {
	cPix   *C.PIX
	closed bool
	lock   sync.Mutex
}

// Deletes the pic, this must be called
func (p *pixStruct) Free() {
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

// ImportFile creates a new pixStruct from given filename
func ImportFile(filename string) (*pixStruct, error) {
	cFilename := C.CString(filename)
	defer C.free(unsafe.Pointer(cFilename))

	// create new PIX
	CPIX := C.pixRead(cFilename)
	if CPIX == nil {
		return nil, errors.New("Unable to read file " + filename)
	}

	// all done
	pix := &pixStruct{
		cPix: CPIX,
	}
	return pix, nil
}

// ImportMem creates a new pixStruct instance from a byte array
func ImportMem(image *[]byte) (*pixStruct, error) {
	ptr := C.uglycast(unsafe.Pointer(&(*image)[0]))
	CPIX := C.pixReadMem(ptr, C.size_t(len(*image)))
	if CPIX == nil {
		return nil, errors.New("Not a valid image file")
	}
	pix := &pixStruct{
		cPix: CPIX,
	}
	return pix, nil
}

// ----------- FUNCTIONS ----------

func (p *pixStruct) pixStructFindSkew() (float32, float32) {
	var angle, conf C.l_float32
	C.pixFindSkew(p.cPix, &angle, &conf)
	return float32(angle), float32(conf)
}
