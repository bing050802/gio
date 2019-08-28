// SPDX-License-Identifier: Unlicense OR MIT

// +build linux,!android

package app

import "unsafe"

func main() {
	<-mainDone
}

func createWindow(window *Window, opts *windowOptions) error {
	return createWindowWayland(window, opts)
}

type window struct {
	wl *wlWindow
}

func (w *window) setAnimating(anim bool) {
	w.wl.setAnimating(anim)
}

func (w *window) display() unsafe.Pointer {
	return w.wl.display()
}

func (w *window) nativeWindow(visID int) (unsafe.Pointer, int, int) {
	return w.wl.nativeWindow(visID)
}

func (w *window) showTextInput(show bool) {
	w.wl.showTextInput(show)
}
