// SPDX-License-Identifier: Unlicense OR MIT

// +build linux,!android,!nox11

package app

import "unsafe"

type eglWindow struct {
	x11 *x11EGLWindow
	wl  *wlEGLWindow
}

func (w *window) newEGLWindow(ew unsafe.Pointer, width, height int) (*eglWindow, error) {
	if w.wl != nil {
		return w.wl.newEGLWindow(ew, width, height)
	}
	return w.x11.newEGLWindow(ew, width, height)
}

func (w *eglWindow) window() unsafe.Pointer {
	if w.wl != nil {
		return w.wl.window()
	}
	return w.x11.window()
}

func (w *eglWindow) resize(width, height int) {
	if w.wl != nil {
		w.wl.resize(width, height)
	} else {
		w.x11.resize(width, height)
	}
}

func (w *eglWindow) destroy() {
	if w.wl != nil {
		w.wl.destroy()
	} else {
		w.x11.destroy()
	}
}
