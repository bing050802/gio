// SPDX-License-Identifier: Unlicense OR MIT

// +build linux windows

package app

import (
	"errors"
	"fmt"
	"runtime"
	"strings"
	"unsafe"

	"gioui.org/ui/app/internal/gl"
)

type context struct {
	c             *gl.Functions
	driver        *window
	eglCtx        *eglContext
	nwindow       _EGLNativeWindowType
	eglWin        *eglWindow
	eglSurf       _EGLSurface
	width, height int
	// For sRGB emulation.
	srgbFBO *gl.SRGBFBO
}

type eglContext struct {
	disp     _EGLDisplay
	config   _EGLConfig
	ctx      _EGLContext
	visualID int
	srgb     bool
}

var (
	nilEGLSurface          _EGLSurface
	nilEGLContext          _EGLContext
	nilEGLConfig           _EGLConfig
	nilEGLNativeWindowType _EGLNativeWindowType
)

const (
	_EGL_ALPHA_SIZE             = 0x3021
	_EGL_BLUE_SIZE              = 0x3022
	_EGL_CONFIG_CAVEAT          = 0x3027
	_EGL_CONTEXT_CLIENT_VERSION = 0x3098
	_EGL_DEPTH_SIZE             = 0x3025
	_EGL_GL_COLORSPACE_KHR      = 0x309d
	_EGL_GL_COLORSPACE_SRGB_KHR = 0x3089
	_EGL_GREEN_SIZE             = 0x3023
	_EGL_EXTENSIONS             = 0x3055
	_EGL_NATIVE_VISUAL_ID       = 0x302e
	_EGL_NONE                   = 0x3038
	_EGL_OPENGL_ES2_BIT         = 0x4
	_EGL_RED_SIZE               = 0x3024
	_EGL_RENDERABLE_TYPE        = 0x3040
	_EGL_SURFACE_TYPE           = 0x3033
	_EGL_WINDOW_BIT             = 0x4
)

func (c *context) Release() {
	if c.srgbFBO != nil {
		c.srgbFBO.Release()
	}
	if c.eglSurf != nilEGLSurface {
		eglMakeCurrent(c.eglCtx.disp, nilEGLSurface, nilEGLSurface, nilEGLContext)
		eglDestroySurface(c.eglCtx.disp, c.eglSurf)
		c.eglSurf = nilEGLSurface
	}
	if c.eglWin != nil {
		c.eglWin.destroy()
		c.eglWin = nil
	}
	if c.eglCtx != nil {
		eglDestroyContext(c.eglCtx.disp, c.eglCtx.ctx)
		eglTerminate(c.eglCtx.disp)
		eglReleaseThread()
		c.eglCtx = nil
	}
	c.driver = nil
}

func (c *context) Present() error {
	if c.eglWin == nil {
		panic("context is not active")
	}
	if c.srgbFBO != nil {
		c.srgbFBO.Blit()
	}
	if !eglSwapBuffers(c.eglCtx.disp, c.eglSurf) {
		return fmt.Errorf("eglSwapBuffers failed (%x)", eglGetError())
	}
	if c.srgbFBO != nil {
		c.srgbFBO.AfterPresent()
	}
	return nil
}

func newContext(w *window) (*context, error) {
	eglCtx, err := createContext(_EGLNativeDisplayType(w.display()))
	if err != nil {
		return nil, err
	}
	c := &context{
		driver: w,
		eglCtx: eglCtx,
		c:      new(gl.Functions),
	}
	return c, nil
}

func (c *context) Functions() *gl.Functions {
	return c.c
}

func (c *context) Lock() {}

func (c *context) Unlock() {}

func (c *context) MakeCurrent() error {
	w, width, height := c.driver.nativeWindow(int(c.eglCtx.visualID))
	win := _EGLNativeWindowType(w)
	if c.nwindow == win && width == c.width && height == c.height {
		return nil
	}
	if win == nilEGLNativeWindowType {
		if c.srgbFBO != nil {
			c.srgbFBO.Release()
			c.srgbFBO = nil
		}
	}
	if c.eglSurf != nilEGLSurface {
		// Make sure any in-flight GL commands are complete.
		c.c.Finish()
		eglMakeCurrent(c.eglCtx.disp, nilEGLSurface, nilEGLSurface, nilEGLContext)
		eglDestroySurface(c.eglCtx.disp, c.eglSurf)
		c.eglSurf = nilEGLSurface
	}
	c.width, c.height = width, height
	c.nwindow = win
	if c.nwindow == nilEGLNativeWindowType {
		if c.eglWin != nil {
			c.eglWin.destroy()
			c.eglWin = nil
		}
		return nil
	}
	if c.eglWin == nil {
		var err error
		c.eglWin, err = c.driver.newEGLWindow(unsafe.Pointer(win), width, height)
		if err != nil {
			return err
		}
	} else {
		c.eglWin.resize(width, height)
	}
	eglSurf, err := createSurfaceAndMakeCurrent(c.eglCtx, _EGLNativeWindowType(c.eglWin.window()))
	c.eglSurf = eglSurf
	if err != nil {
		c.eglWin.destroy()
		c.eglWin = nil
		c.nwindow = nilEGLNativeWindowType
		return err
	}
	if c.eglCtx.srgb {
		return nil
	}
	if c.srgbFBO == nil {
		var err error
		c.srgbFBO, err = gl.NewSRGBFBO(c.c)
		if err != nil {
			c.Release()
			return err
		}
	}
	if err := c.srgbFBO.Refresh(c.width, c.height); err != nil {
		c.Release()
		return err
	}
	return nil
}

func hasExtension(exts []string, ext string) bool {
	for _, e := range exts {
		if ext == e {
			return true
		}
	}
	return false
}

func createContext(disp _EGLNativeDisplayType) (*eglContext, error) {
	eglDisp := eglGetDisplay(disp)
	if eglDisp == 0 {
		return nil, fmt.Errorf("eglGetDisplay(_EGL_DEFAULT_DISPLAY) failed: 0x%x", eglGetError())
	}
	major, minor, ret := eglInitialize(eglDisp)
	if !ret {
		return nil, fmt.Errorf("eglInitialize failed: 0x%x", eglGetError())
	}
	// sRGB framebuffer support on EGL 1.5 or if EGL_KHR_gl_colorspace is supported.
	exts := strings.Split(eglQueryString(eglDisp, _EGL_EXTENSIONS), " ")
	srgb := major > 1 || minor >= 5 || hasExtension(exts, "EGL_KHR_gl_colorspace")
	attribs := []_EGLint{
		_EGL_RENDERABLE_TYPE, _EGL_OPENGL_ES2_BIT,
		_EGL_SURFACE_TYPE, _EGL_WINDOW_BIT,
		_EGL_BLUE_SIZE, 8,
		_EGL_GREEN_SIZE, 8,
		_EGL_RED_SIZE, 8,
		_EGL_CONFIG_CAVEAT, _EGL_NONE,
	}
	if srgb {
		if runtime.GOOS == "linux" {
			// Some Mesa drivers crash if an sRGB framebuffer is requested without alpha.
			// https://bugs.freedesktop.org/show_bug.cgi?id=107782.
			attribs = append(attribs, _EGL_ALPHA_SIZE, 1)
		}
		// Only request a depth buffer if we're going to render directly to the framebuffer.
		attribs = append(attribs, _EGL_DEPTH_SIZE, 16)
	}
	attribs = append(attribs, _EGL_NONE)
	eglCfg, ret := eglChooseConfig(eglDisp, attribs)
	if !ret {
		return nil, fmt.Errorf("eglChooseConfig failed: 0x%x", eglGetError())
	}
	if eglCfg == nilEGLConfig {
		return nil, errors.New("eglChooseConfig returned 0 configs")
	}
	var eglCtx _EGLContext
	ctxAttribs := []_EGLint{
		_EGL_CONTEXT_CLIENT_VERSION, 3,
		_EGL_NONE,
	}
	eglCtx = eglCreateContext(eglDisp, eglCfg, nilEGLContext, ctxAttribs)
	if eglCtx == nilEGLContext {
		return nil, fmt.Errorf("eglCreateContext failed: 0x%x", eglGetError())
	}
	visID, ret := eglGetConfigAttrib(eglDisp, eglCfg, _EGL_NATIVE_VISUAL_ID)
	if !ret {
		return nil, errors.New("newContext: eglGetConfigAttrib for _EGL_NATIVE_VISUAL_ID failed")
	}
	return &eglContext{
		disp:     eglDisp,
		config:   _EGLConfig(eglCfg),
		ctx:      _EGLContext(eglCtx),
		visualID: int(visID),
		srgb:     srgb,
	}, nil
}

func createSurfaceAndMakeCurrent(eglCtx *eglContext, win _EGLNativeWindowType) (_EGLSurface, error) {
	var surfAttribs []_EGLint
	if eglCtx.srgb {
		surfAttribs = append(surfAttribs, _EGL_GL_COLORSPACE_KHR, _EGL_GL_COLORSPACE_SRGB_KHR)
	}
	surfAttribs = append(surfAttribs, _EGL_NONE)
	eglSurf := eglCreateWindowSurface(eglCtx.disp, eglCtx.config, win, surfAttribs)
	if eglSurf == nilEGLSurface {
		return nilEGLSurface, fmt.Errorf("newContext: eglCreateWindowSurface failed 0x%x", eglGetError())
	}
	if !eglMakeCurrent(eglCtx.disp, eglSurf, eglSurf, eglCtx.ctx) {
		eglDestroySurface(eglCtx.disp, eglSurf)
		return nilEGLSurface, fmt.Errorf("eglMakeCurrent error 0x%x", eglGetError())
	}
	// eglSwapInterval 1 leads to erratic frame rates and unnecessary blocking.
	// We rely on platform specific frame rate limiting instead, except on Windows
	// where eglSwapInterval is all there is.
	if runtime.GOOS != "windows" {
		eglSwapInterval(eglCtx.disp, 0)
	} else {
		eglSwapInterval(eglCtx.disp, 1)
	}
	return eglSurf, nil
}
