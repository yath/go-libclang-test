package main

/*
#cgo LDFLAGS: -L/usr/lib/llvm-7/lib -lclang
#cgo CFLAGS: -I/usr/lib/llvm-7/include

#include <stdlib.h>
#include <clang-c/Index.h>

extern void parseAndVisit(const char *, void *);


*/
import "C"

import (
	"fmt"
	"runtime"
	"unsafe"
)

type ctype string

type funArg struct {
	name string
	t    ctype
}

type location struct {
	file         string
	line, column uint
}

type funDecl struct {
	name string
	loc  location
	rett ctype
	args []funArg
}

func cxstring(s C.CXString) string {
	defer C.clang_disposeString(s)
	return C.GoString(C.clang_getCString(s))
}

func cxtype(s C.CXType) ctype {
	return ctype(cxstring(C.clang_getTypeSpelling(s)))
}

type visitorCallback func(funDecl)

var visitorCallbacks = make(map[unsafe.Pointer]visitorCallback)

func addVisitorCallback(cb visitorCallback) unsafe.Pointer {
	key := unsafe.Pointer(C.CString("x"))
	runtime.SetFinalizer(&key, func(w *unsafe.Pointer) { C.free(*w) })
	visitorCallbacks[key] = cb
	return key
}

func visitFunctions(ppHeader string, v visitorCallback) {
	k := addVisitorCallback(v)
	defer delete(visitorCallbacks, k)

	fn := C.CString(ppHeader)
	defer C.free(unsafe.Pointer(fn))

	C.parseAndVisit(fn, k)
}

//export goVisitor
func goVisitor(c, parent C.CXCursor, cd C.CXClientData) C.enum_CXChildVisitResult {
	v, ok := visitorCallbacks[unsafe.Pointer(cd)]
	if !ok {
		panic(fmt.Sprintf("goVisitor called for unknown callback %v", cd))
	}

	_ = parent
	t := C.clang_getCursorType(c)
	if c.kind != C.CXCursor_FunctionDecl {
		return C.CXChildVisit_Recurse
	}

	if C.clang_getCursorVisibility(c) != C.CXVisibility_Default {
		return C.CXChildVisit_Recurse
	}

	var cfile C.CXString
	var cline, ccol C.unsigned
	C.clang_getPresumedLocation(C.clang_getCursorLocation(c), &cfile, &cline, &ccol)

	f := funDecl{
		name: cxstring(C.clang_getCursorSpelling(c)),
		rett: cxtype(C.clang_getResultType(t)),
		loc: location{
			file:   cxstring(cfile),
			line:   uint(cline),
			column: uint(ccol),
		},
	}

	for i := 0; i < int(C.clang_Cursor_getNumArguments(c)); i++ {
		argc := C.clang_Cursor_getArgument(c, C.uint(i))
		arg := funArg{
			t:    cxtype(C.clang_getCursorType(argc)),
			name: cxstring(C.clang_getCursorSpelling(argc)),
		}
		f.args = append(f.args, arg)
	}

	v(f)

	return C.CXChildVisit_Recurse
}

func getFunDecls(ppHeader string) []funDecl {
	ret := []funDecl{}
	visitFunctions(ppHeader, func(f funDecl) { ret = append(ret, f) })
	return ret
}
