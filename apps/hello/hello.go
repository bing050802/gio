// SPDX-License-Identifier: Unlicense OR MIT

package main

// A simple Gio program. See https://gioui.org for more information.

import (
	"image/color"
	"log"

	"gioui.org/ui"
	"gioui.org/ui/app"
	"gioui.org/ui/layout"
	"gioui.org/ui/measure"
	"gioui.org/ui/paint"
	"gioui.org/ui/text"

	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/font/sfnt"
)

func main() {
	go func() {
		w := app.NewWindow()
		if err := loop(w); err != nil {
			log.Fatal(err)
		}
	}()
	app.Main()
}

func loop(w *app.Window) error {
	regular, err := sfnt.Parse(goregular.TTF)
	if err != nil {
		panic("failed to load font")
	}
	var cfg app.Config
	var faces measure.Faces
	maroon := color.RGBA{127, 0, 0, 255}
	face := faces.For(regular, ui.Sp(72))
	message := "Hello, Gio"
	ops := new(ui.Ops)
	for {
		e := <-w.Events()
		switch e := e.(type) {
		case app.DestroyEvent:
			return e.Err
		case app.UpdateEvent:
			cfg = e.Config
			faces.Reset(&cfg)
			cs := layout.RigidConstraints(e.Size)
			ops.Reset()
			var material ui.MacroOp
			material.Record(ops)
			paint.ColorOp{Color: maroon}.Add(ops)
			material.Stop()
			text.Label{Material: material, Face: face, Alignment: text.Middle, Text: message}.Layout(ops, cs)
			w.Update(ops)
		}
	}
}
