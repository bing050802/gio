// SPDX-License-Identifier: Unlicense OR MIT

// +build linux,!android,nox11

package app

import "unsafe"

type eglWindow struct {
	wl *wlEGLWindow
}

func (w *window) newEGLWindow(ew unsafe.Pointer, width, height int) (*eglWindow, error) {
	return w.wl.newEGLWindow(ew, width, height)
}

func (w *eglWindow) window() unsafe.Pointer {
	return w.wl.window()
}

func (w *eglWindow) resize(width, height int) {
	w.wl.resize(width, height)
}

func (w *eglWindow) destroy() {
	w.wl.destroy()
}
