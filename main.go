package main

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"gioui.org/app"
	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

type C = layout.Context
type D = layout.Dimensions

var boilDurationInput widget.Editor // boilDurationInput is a textfield to input boil duration

var boiling bool
var boilDuration float32

var progress float32
var progressIncrementer chan float32 // progressIncrementer is the channel into which we send values, in this case of type float32

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

				if progress >= 1 {
					progress = 0
				}

				inputString := boilDurationInput.Text()
				inputString = strings.TrimSpace(inputString)
				inputFloat, _ := strconv.ParseFloat(inputString, 32)
				boilDuration = float32(inputFloat)
				boilDuration = boilDuration / (1 - progress)
			}

			layout.Flex{
				// vertical alignment, from top to bottom
				Axis: layout.Vertical,
				// Empty space is left at the start, i.e. at the top
				Spacing: layout.SpaceStart,
			}.Layout(gtx,
				// the egg
				layout.Rigid(
					func(gtx C) D {
						// Draw a custom path, shaped like an egg
						var eggPath clip.Path
						op.Offset(image.Pt(gtx.Dp(200), gtx.Dp(125))).Add(gtx.Ops)
						eggPath.Begin(gtx.Ops)
						// rotate from 0 to 360 degrees
						for deg := 0.0; deg <= 360; deg++ {
							// Egg math: https://observablehq.com/@toja/egg-curve
							// convert degrees to radians
							rad := deg / 360 * 2 * math.Pi
							// Trig gives the distance in X and Y direction
							cosT := math.Cos(rad)
							sinT := math.Sin(rad)
							// Constants to define the egg shape
							a := 110.0
							b := 150.0
							d := 20.0
							// the x/y coordinates
							x := a * cosT
							y := -(math.Sqrt(b*b-d*d*cosT*cosT) + d*sinT) * sinT
							// Finally the point on the outline
							p := f32.Pt(float32(x), float32(y))
							// Draw the line to this point
							eggPath.LineTo(p)
						}
						// close the path
						eggPath.Close()

						// get hold of the actual clip
						eggArea := clip.Outline{Path: eggPath.End()}.Op()

						// Fill the shape
						// color := color.NRGBA{R: 255, G: 239, B: 174, A: 255}
						color := color.NRGBA{R: 255, G: uint8(239 * (1 - progress)), B: uint8(174 * (1 - progress)), A: 255}
						paint.FillShape(gtx.Ops, color, eggArea)

						d := image.Point{Y: 375}
						return layout.Dimensions{Size: d}
					},
				),
				// the inputbox
				layout.Rigid(
					func(gtx C) D {
						// Wrap the editor in material design
						ed := material.Editor(th, &boilDurationInput, "sec")

						// Define characteristics of the input box
						boilDurationInput.SingleLine = true
						boilDurationInput.Alignment = text.Middle

						if boiling && progress < 1 {
							boilRemain := (1 - progress) * boilDuration
							// Format to 1 decimal
							inputStr := fmt.Sprintf("%.1f", math.Round(float64(boilRemain)*10)/10)
							// Update the text in the inputbox
							boilDurationInput.SetText(inputStr)
						}
						// Define insets ...
						margins := layout.Inset{
							Top:    unit.Dp(0),
							Right:  unit.Dp(170),
							Bottom: unit.Dp(40),
							Left:   unit.Dp(170),
						}
						// ... and borders ...
						border := widget.Border{
							Color:        color.NRGBA{R: 204, G: 204, B: 204, A: 255},
							CornerRadius: unit.Dp(3),
							Width:        unit.Dp(2),
						}
						// ... before laying it out, one inside the other
						return margins.Layout(gtx,
							func(gtx C) D {
								return border.Layout(gtx, ed.Layout)
							},
						)
					},
				),
				// the progressbar
				layout.Rigid(
					func(gtx C) D {
						bar := material.ProgressBar(th, progress) // use the global progress variable
						return bar.Layout(gtx)
					},
				),
				// the button
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
								text = "Start"
								if boiling && progress < 1 {
									text = "Stop"
								}
								if boiling && progress >= 1 {
									text = "Finished"
								}
								btn := material.Button(th, &startButton, text) // define button
								return btn.Layout(gtx)
							},
						)
					},
				),
				// the circle
				// layout.Rigid(
				// 	func(gtx C) D {
				// 	  circle := clip.Ellipse{
				// 		 // Hard coding the x coordinate. Try resizing the window
				// 		 // Min: image.Pt(80, 0),
				// 		 // Max: image.Pt(320, 240),
				// 		 // Soft coding the x coordinate. Try resizing the window
				// 		 Min: image.Pt(gtx.Constraints.Max.X/2 - 120, 0),
				// 		 Max: image.Pt(gtx.Constraints.Max.X/2 + 120, 240),
				// 	  }.Op(gtx.Ops)
				// 	  color := color.NRGBA{G: 200, A: 255}
				// 	  paint.FillShape(gtx.Ops, color, circle)
				// 	  d := image.Point{Y: 400}
				// 	  return layout.Dimensions{Size: d}
				// 	},
				//   ),
			)
			typ.Frame(gtx.Ops) // send the operations from the context gtx to the FrameEvent

		// and this is sent when the app should exit
		case app.DestroyEvent:
			return errors.New("user exited the application")
		}
	}
}
