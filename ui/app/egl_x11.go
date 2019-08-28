// SPDX-License-Identifier: Unlicense OR MIT

// +build linux,!android

package app

/*
#include <X11/Xlib.h>
#include <X11/Xatom.h>
#include <X11/Xutil.h>
*/
import "C"
import "unsafe"

type x11EGLWindow struct {
	w  *x11Window
	xw unsafe.Pointer
}

func (w *x11Window) newEGLWindow(xw unsafe.Pointer, width, height int) (*eglWindow, error) {
	return &eglWindow{x11: &x11EGLWindow{w: w, xw: xw}}, nil
}

func (w *x11EGLWindow) window() unsafe.Pointer {
	return w.xw
}

func (w *x11EGLWindow) resize(width, height int) {
	var change C.XWindowChanges
	change.width = C.int(width)
	change.height = C.int(height)
	C.XConfigureWindow(w.w.x, w.w.xw, C.CWWidth|C.CWHeight, &change)
}

func (w *x11EGLWindow) destroy() {
	// destroyed by x11Window.destroy
}
