package main

import (
	"errors"
	"log"
	"os"
	"time"

	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

type C = layout.Context
type D = layout.Dimensions

var progress float32
var progressIncrementer chan float32 // progressIncrementer is the channel into which we send values, in this case of type float32
var boiling bool

func main() {
	// setup a separate channel to provide ticks to increment progress
	progressIncrementer = make(chan float32)
	go func() {
		for { // every 1/25th of a second the number 0.004 is injected into the channel
			time.Sleep(time.Second / 25)
			progressIncrementer <- 0.004
		}
	}()
	go func() {
		// create new window
		w := new(app.Window)
		w.Option(app.Title("Egg timer"), app.Size(unit.Dp(400), unit.Dp(600)))

		if err := draw(w); err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}()
	app.Main()
}

func draw(w *app.Window) error {
	// ops are the OPERATIONS from the UI
	var ops op.Ops

	// startButton is a clickable widget
	var startButton widget.Clickable

	// th defines the material design style
	th := material.NewTheme()

	// Start the goroutine that listens for events in the incrementer channel
	go func() {
		for p := range progressIncrementer {
			if boiling && progress < 1 {
				progress += p
				w.Invalidate()
			}
		}
	}()

	// listen for events in the window 	(this is the EVENT LOOP)
	for {
		// first grab the event
		evt := w.Event()

		// then detect the type  (this is a TYPE SWITCH)
		switch typ := evt.(type) {

		// this is sent when the app should re-render.
		case app.FrameEvent:
			gtx := app.NewContext(&ops, typ) // define a new GRAPHICAL CONTEXT (gtx)

			if startButton.Clicked(gtx) {
				boiling = !boiling
			}

			layout.Flex{
				// vertical alignment, from top to bottom
				Axis: layout.Vertical,
				// Empty space is left at the start, i.e. at the top
				Spacing: layout.SpaceStart,
			}.Layout(gtx,
				layout.Rigid( // Rigid() accepts a WIDGET. A widget is simply something that returns its own DIMENSIONS.
					func(gtx C) D {
						// ONE: First define margins around the button using layout.Inset ...
						margins := layout.Inset{
							Top:    unit.Dp(25),
							Bottom: unit.Dp(25),
							Left:   unit.Dp(35),
							Right:  unit.Dp(35),
						}
						// marginsAutoSpacedEvenly := layout.UniformInset(25)	// we can also do this!
						// TWO: ... then we lay out those margins ...
						return margins.Layout(gtx,
							// THREE: ... and finally within the margins, we define and lay out the button
							func(gtx C) D {
								var text string
								if !boiling {
									text = "Start"
								} else {
									text = "Stop"
								}
								btn := material.Button(th, &startButton, text) // define button
								return btn.Layout(gtx)
							},
						)
					},
				),
				layout.Rigid(
					func(gtx C) D {
						bar := material.ProgressBar(th, progress) // use the global progress variable
						return bar.Layout(gtx)
					},
				),
			)

			typ.Frame(gtx.Ops) // send the operations from the context gtx to the FrameEvent

		// and this is sent when the app should exit
		case app.DestroyEvent:
			return errors.New("user exited the application")
		}
	}
}
