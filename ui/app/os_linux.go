// SPDX-License-Identifier: Unlicense OR MIT

// +build linux,!android

package app

import "unsafe"

func main() {
	<-mainDone
}

func createWindow(window *Window, opts *windowOptions) error {
	err := createWindowWayland(window, opts)
	if err == errWLDisplayConnectFailed {
		return createWindowX11(window, opts)
	}
	return err
}

type window struct {
	x11 *x11Display
	wl  *wlWindow
}

func (w *window) setAnimating(anim bool) {
	if w.wl != nil {
		w.wl.setAnimating(anim)
	} else {
		w.x11.setAnimating(anim)
	}
}

func (w *window) display() unsafe.Pointer {
	if w.wl != nil {
		return w.wl.display()
	}
	return w.x11.display()
}

func (w *window) nativeWindow(visID int) (unsafe.Pointer, int, int) {
	if w.wl != nil {
		return w.wl.nativeWindow(visID)
	}
	return w.x11.nativeWindow(visID)
}

func (w *window) showTextInput(show bool) {
	if w.wl != nil {
		w.wl.showTextInput(show)
	} else {
		w.x11.showTextInput(show)
	}
}
