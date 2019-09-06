// SPDX-License-Identifier: Unlicense OR MIT

package app

/*
#cgo LDFLAGS: -lEGL

#include <EGL/egl.h>
#include <EGL/eglext.h>
#include <GLES2/gl2.h>
#include <GLES3/gl3.h>
*/
import "C"
import (
	"unsafe"
)

type (
	_EGLint     = C.EGLint
	_EGLDisplay = C.EGLDisplay
	_EGLConfig  = C.EGLConfig
	_EGLContext = C.EGLContext
	_EGLSurface = C.EGLSurface
)

func eglChooseConfig(disp _EGLDisplay, attribs []_EGLint) (_EGLConfig, bool) {
	var cfg C.EGLConfig
	var ncfg C.EGLint
	if C.eglChooseConfig(disp, &attribs[0], &cfg, 1, &ncfg) != C.EGL_TRUE {
		return nil, false
	}
	return _EGLConfig(cfg), true
}

func eglCreateContext(disp _EGLDisplay, cfg _EGLConfig, shareCtx _EGLContext, attribs []_EGLint) _EGLContext {
	ctx := C.eglCreateContext(disp, cfg, shareCtx, &attribs[0])
	return _EGLContext(ctx)
}

func eglDestroySurface(disp _EGLDisplay, surf _EGLSurface) bool {
	return C.eglDestroySurface(disp, surf) == C.EGL_TRUE
}

func eglDestroyContext(disp _EGLDisplay, ctx _EGLContext) bool {
	return C.eglDestroyContext(disp, ctx) == C.EGL_TRUE
}

func eglGetConfigAttrib(disp _EGLDisplay, cfg _EGLConfig, attr _EGLint) (_EGLint, bool) {
	var val _EGLint
	ret := C.eglGetConfigAttrib(disp, cfg, attr, &val)
	return val, ret == C.EGL_TRUE
}

func eglGetError() _EGLint {
	return C.eglGetError()
}

func eglInitialize(disp _EGLDisplay) (_EGLint, _EGLint, bool) {
	var maj, min _EGLint
	ret := C.eglInitialize(disp, &maj, &min)
	return maj, min, ret == C.EGL_TRUE
}

func eglMakeCurrent(disp _EGLDisplay, draw, read _EGLSurface, ctx _EGLContext) bool {
	return C.eglMakeCurrent(disp, draw, read, ctx) == C.EGL_TRUE
}

func eglReleaseThread() bool {
	return C.eglReleaseThread() == C.EGL_TRUE
}

func eglSwapBuffers(disp _EGLDisplay, surf _EGLSurface) bool {
	return C.eglSwapBuffers(disp, surf) == C.EGL_TRUE
}

func eglSwapInterval(disp _EGLDisplay, interval _EGLint) bool {
	return C.eglSwapInterval(disp, interval) == C.EGL_TRUE
}

func eglTerminate(disp _EGLDisplay) bool {
	return C.eglTerminate(disp) == C.EGL_TRUE
}

func eglQueryString(disp _EGLDisplay, name _EGLint) string {
	return C.GoString(C.eglQueryString(disp, name))
}

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
