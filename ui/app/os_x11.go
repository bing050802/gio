// SPDX-License-Identifier: Unlicense OR MIT

// +build linux,!android

package app

/*
#cgo LDFLAGS: -lX11

#include <stdlib.h>
#include <X11/Xlib.h>
#include <X11/Xatom.h>
#include <X11/Xutil.h>
#include <X11/XKBlib.h>

#define GIO_FIELD_OFFSET(typ, field) const int gio_##typ##_##field##_off = offsetof(typ, field)

GIO_FIELD_OFFSET(XClientMessageEvent, data);
GIO_FIELD_OFFSET(XConfigureEvent, width);
GIO_FIELD_OFFSET(XConfigureEvent, height);
GIO_FIELD_OFFSET(XButtonEvent, x);
GIO_FIELD_OFFSET(XButtonEvent, y);
GIO_FIELD_OFFSET(XButtonEvent, state);
GIO_FIELD_OFFSET(XButtonEvent, button);
GIO_FIELD_OFFSET(XButtonEvent, time);
GIO_FIELD_OFFSET(XMotionEvent, x);
GIO_FIELD_OFFSET(XMotionEvent, y);
GIO_FIELD_OFFSET(XMotionEvent, time);
GIO_FIELD_OFFSET(XkbAnyEvent, xkb_type);
GIO_FIELD_OFFSET(XkbAnyEvent, time);
GIO_FIELD_OFFSET(XkbStateNotifyEvent, keycode);
GIO_FIELD_OFFSET(XkbStateNotifyEvent, event_type);
GIO_FIELD_OFFSET(XkbStateNotifyEvent, req_major);
GIO_FIELD_OFFSET(XkbStateNotifyEvent, req_minor);
*/
import "C"
import (
	"errors"
	"image"
	"sync"
	"time"
	"unsafe"

	"gioui.org/ui/f32"
	"gioui.org/ui/pointer"
)

type x11Window struct {
	w  *Window
	x  *C.Display
	xw C.Window

	evDelWindow C.Atom

	width  int
	height int
	cfg    Config

	xkb struct {
		opcode  C.int
		event   C.int
		errcode C.int
		major   C.int
		minor   C.int
	}
}

func (w *x11Window) setAnimating(anim bool) {
	// TODO(dennwc): implement animation state
}

func (w *x11Window) showTextInput(show bool) {}

func (w *x11Window) display() unsafe.Pointer {
	// TODO(dennwc): We have an awesome X library written in pure Go, however,
	//               we can't use it because of this specific function.
	//               The *C.Display pointer is required to call eglGetDisplay,
	//               so we can't really implement the call in pure Go.
	//               Thus, we have to use Xlib for everything.
	return unsafe.Pointer(w.x)
}

func (w *x11Window) nativeWindow(visID int) (unsafe.Pointer, int, int) {
	return unsafe.Pointer(uintptr(w.xw)), w.width, w.height
}

func (w *x11Window) setStage(s Stage) {
	w.w.event(StageEvent{s})
}

func (w *x11Window) loop() {
	for {
		var xev xEvent
		C.XNextEvent(w.x, (*C.XEvent)(unsafe.Pointer(&xev)))
		switch xev.Type {
		case C.ButtonPress, C.ButtonRelease:
			ev := pointer.Event{
				Type:   pointer.Press,
				Source: pointer.Mouse,
				Position: f32.Point{
					X: float32(xev.GetButtonX()),
					Y: float32(xev.GetButtonY()),
				},
				Time: xev.GetButtonTime(),
			}
			if xev.Type == C.ButtonRelease {
				ev.Type = pointer.Release
			}
			const scrollScale = 10
			switch xev.GetButtonButton() {
			case C.Button1:
				// left click by default
			case C.Button4:
				// scroll up
				ev.Type = pointer.Move
				ev.Scroll.Y = -scrollScale
			case C.Button5:
				// scroll down
				ev.Type = pointer.Move
				ev.Scroll.Y = +scrollScale
			default:
				continue
			}
			w.w.event(ev)
			w.draw()
		case C.MotionNotify:
			w.w.event(pointer.Event{
				Type:   pointer.Move,
				Source: pointer.Mouse,
				Position: f32.Point{
					X: float32(xev.GetMotionX()),
					Y: float32(xev.GetMotionY()),
				},
				Time: xev.GetMotionTime(),
			})
			w.draw()
		case C.Expose: // update
			w.draw()
		case C.ConfigureNotify: // window configuration change
			oldW, oldH := w.width, w.height
			w.width = int(xev.GetConfigureWidth())
			w.height = int(xev.GetConfigureHeight())
			if oldW != w.width || oldH != w.height {
				w.draw()
			}
		case C.ClientMessage: // extensions
			switch xev.GetClientDataLong()[0] {
			case C.long(w.evDelWindow):
				return
			}
		case C.KeyPress, C.KeyRelease:
			// TODO(dennwc): keyboard press
		case w.xkb.event:
			switch xev.GetXkbType() {
			// TODO(dennwc): Xkb state
			}
		}
	}
}

func (w *x11Window) destroy() {
	C.XDestroyWindow(w.x, w.xw)
	C.XCloseDisplay(w.x)
}

func (w *x11Window) draw() {
	w.w.event(UpdateEvent{
		Size: image.Point{
			X: w.width,
			Y: w.height,
		},
		Config: w.cfg,
		sync:   false,
	})
}

const xEventSize = unsafe.Sizeof(C.XEvent{})

// Make sure the Go struct has the same size.
// We can't use C.XEvent directly because it's a union.
var _ = [1]struct{}{}[unsafe.Sizeof(xEvent{})-xEventSize]

type xEvent struct {
	Type C.int
	Data [xEventSize - unsafe.Sizeof(C.int(0))]byte
}

func (e *xEvent) getInt(off int) C.int {
	return *(*C.int)(unsafe.Pointer(uintptr(unsafe.Pointer(e)) + uintptr(off)))
}

func (e *xEvent) getUint(off int) C.uint {
	return *(*C.uint)(unsafe.Pointer(uintptr(unsafe.Pointer(e)) + uintptr(off)))
}

func (e *xEvent) getUlong(off int) C.ulong {
	return *(*C.ulong)(unsafe.Pointer(uintptr(unsafe.Pointer(e)) + uintptr(off)))
}

func (e *xEvent) getUlongMs(off int) time.Duration {
	return time.Duration(e.getUlong(off)) * time.Millisecond
}

// GetConfigureWidth returns a XEvent.xconfigure.width field.
func (e *xEvent) GetConfigureWidth() C.int {
	return e.getInt(int(C.gio_XConfigureEvent_width_off))
}

// GetConfigureWidth returns a XEvent.xconfigure.height field.
func (e *xEvent) GetConfigureHeight() C.int {
	return e.getInt(int(C.gio_XConfigureEvent_height_off))
}

// GetButtonX returns a XEvent.xbutton.x field.
func (e *xEvent) GetButtonX() C.int {
	return e.getInt(int(C.gio_XButtonEvent_x_off))
}

// GetButtonY returns a XEvent.xbutton.y field.
func (e *xEvent) GetButtonY() C.int {
	return e.getInt(int(C.gio_XButtonEvent_y_off))
}

// GetButtonState returns a XEvent.xbutton.state field.
func (e *xEvent) GetButtonState() C.uint {
	return e.getUint(int(C.gio_XButtonEvent_state_off))
}

// GetButtonButton returns a XEvent.xbutton.button field.
func (e *xEvent) GetButtonButton() C.uint {
	return e.getUint(int(C.gio_XButtonEvent_button_off))
}

// GetButtonTime returns a XEvent.xbutton.time field.
func (e *xEvent) GetButtonTime() time.Duration {
	return e.getUlongMs(int(C.gio_XButtonEvent_time_off))
}

// GetMotionX returns a XEvent.xmotion.x field.
func (e *xEvent) GetMotionX() C.int {
	return e.getInt(int(C.gio_XMotionEvent_x_off))
}

// GetMotionY returns a XEvent.xmotion.y field.
func (e *xEvent) GetMotionY() C.int {
	return e.getInt(int(C.gio_XMotionEvent_y_off))
}

// GetMotionTime returns a XEvent.xmotion.time field.
func (e *xEvent) GetMotionTime() time.Duration {
	return e.getUlongMs(int(C.gio_XMotionEvent_time_off))
}

// GetClientDataLong returns a XEvent.xclient.data.l field.
func (e *xEvent) GetClientDataLong() [5]C.long {
	ptr := (*[5]C.long)(unsafe.Pointer(uintptr(unsafe.Pointer(e)) + uintptr(C.gio_XClientMessageEvent_data_off)))
	return *ptr
}

// GetXkbType returns a XkbEvent.any.xkb_type field.
func (e *xEvent) GetXkbType() C.int {
	return e.getInt(int(C.gio_XkbAnyEvent_xkb_type_off))
}

// GetXkbTime returns a XkbEvent.any.time field.
func (e *xEvent) GetXkbTime() time.Duration {
	return e.getUlongMs(int(C.gio_XkbAnyEvent_time_off))
}

var (
	x11Threads sync.Once
)

func createWindowX11(w *Window, opts *windowOptions) error {
	var err error
	x11Threads.Do(func() {
		if C.XInitThreads() == 0 {
			err = errors.New("x11: threads init failed")
		}
	})
	if err != nil {
		return err
	}
	disp := C.XOpenDisplay(nil)
	if disp == nil {
		return errors.New("x11: cannot connect to the X server")
	}
	xw := &x11Window{
		w: w, x: disp,
		cfg: Config{pxPerDp: 1, pxPerSp: 1}, // TODO(dennwc): real config
	}
	if C.XkbQueryExtension(disp, &xw.xkb.opcode, &xw.xkb.event, &xw.xkb.opcode, &xw.xkb.major, &xw.xkb.minor) == 0 {
		C.XCloseDisplay(disp)
		return errors.New("x11: Xkb is not supported")
	}
	C.XkbSelectEvents(disp, C.XkbUseCoreKbd, C.XkbAllEventsMask, C.XkbAllEventsMask)

	root := C.XDefaultRootWindow(disp)

	var swa C.XSetWindowAttributes
	swa.event_mask = C.ExposureMask | // update
		C.KeyPressMask | C.KeyReleaseMask | // keyboard
		C.ButtonPressMask | C.ButtonReleaseMask | // mouse clicks
		C.PointerMotionMask | // mouse movement
		C.StructureNotifyMask // resize

	xw.width, xw.height = xw.cfg.Px(opts.Width), xw.cfg.Px(opts.Width)
	xw.xw = C.XCreateWindow(disp, root,
		0, 0, C.uint(xw.width), C.uint(xw.height), 0,
		C.CopyFromParent, C.InputOutput,
		nil, C.CWEventMask|C.CWBackPixel,
		&swa,
	)

	var xattr C.XSetWindowAttributes
	xattr.override_redirect = C.False
	C.XChangeWindowAttributes(disp, xw.xw, C.CWOverrideRedirect, &xattr)

	var hints C.XWMHints
	hints.input = C.True
	hints.flags = C.InputHint
	C.XSetWMHints(disp, xw.xw, &hints)

	// make the window visible on the screen
	C.XMapWindow(disp, xw.xw)

	// set the name
	ctitle := C.CString(opts.Title)
	C.XStoreName(disp, xw.xw, ctitle)
	C.free(unsafe.Pointer(ctitle))

	// extensions
	ckey := C.CString("WM_DELETE_WINDOW")
	xw.evDelWindow = C.XInternAtom(disp, ckey, C.False)
	C.free(unsafe.Pointer(ckey))
	C.XSetWMProtocols(disp, xw.xw, &xw.evDelWindow, 1)

	go func() {
		xw.w.setDriver(&window{x11: xw})
		xw.setStage(StageRunning)
		xw.loop()
		xw.destroy()
		close(mainDone)
	}()
	return nil
}
