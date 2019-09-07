// SPDX-License-Identifier: Unlicense OR MIT

// +build linux,!android,!nox11

package app

/*
#include <X11/Xlib.h>
#include <X11/Xatom.h>
#include <X11/Xutil.h>
*/
import "C"
import "unsafe"

type x11EGLWindow struct {
	w *x11Window
}

func (d *x11Display) newEGLWindow(xw unsafe.Pointer, width, height int) (*eglWindow, error) {
	return &eglWindow{x11: &x11EGLWindow{w: d.mainWin}}, nil
}

func (w *x11EGLWindow) window() unsafe.Pointer {
	return unsafe.Pointer(uintptr(w.w.w))
}

func (w *x11EGLWindow) resize(width, height int) {
	w.w.Resize(width, height)
}

func (w *x11EGLWindow) destroy() {
	// destroyed by x11Display.destroy
}
