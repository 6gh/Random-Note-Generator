package main

import (
	"errors"
	"fmt"
	"image/color"
	"net/url"
	"path"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func createGUI() {
	app := app.NewWithID("xyz.6gh.randomnotegen")

	window := app.NewWindow("Random Note Generator")
	window.Resize(fyne.NewSize(800, 600))

	// help bar at the bottom
	HelpBar := widget.NewToolbar(
		widget.NewToolbarAction(theme.HelpIcon(), func() {
			icon := canvas.NewImageFromResource(window.Icon())
			icon.SetMinSize(fyne.NewSize(128, 128))
			icon.FillMode = canvas.ImageFillContain

			title := canvas.NewText("Random Note Generator", color.White)
			title.TextStyle = fyne.TextStyle{
				Bold: true,
			}
			title.Alignment = fyne.TextAlignCenter
			title.TextSize = 24

			version := canvas.NewText(fyne.CurrentApp().Metadata().Version, color.White)
			version.Alignment = fyne.TextAlignCenter
			version.TextSize = 16

			creator := canvas.NewText("Created by 6gh", color.White)
			creator.Alignment = fyne.TextAlignCenter

			fyneUrl, err := url.Parse("https://fyne.io/")
			handleErr(err)
			fyneLbl := widget.NewHyperlink("Made with Fyne", fyneUrl)
			fyneLbl.Alignment = fyne.TextAlignCenter

			repoUrl, err := url.Parse("https://github.com/6gh/Random-Note-Generator")
			handleErr(err)
			repoLbl := widget.NewHyperlink("Check on GitHub", repoUrl)
			repoLbl.Alignment = fyne.TextAlignCenter

			hBox := container.New(layout.NewGridLayout(2), fyneLbl, repoLbl)
			vBox := container.New(layout.NewVBoxLayout(), icon, title, version, creator, hBox)

			dialog.ShowCustom("About", "Close", vBox, window)
		}),
		widget.NewToolbarAction(theme.SettingsIcon(), func() {
			title := canvas.NewText("Additional Settings", color.White)
			title.TextStyle = fyne.TextStyle{
				Bold: true,
			}
			title.Alignment = fyne.TextAlignCenter
			title.TextSize = 24

			// max notes per track
			MaxNotesTXT := createNumberInput(0, -1)

			// length type
			// ticks, seconds, bars
			LengthTXT := widget.NewSelect([]string{"MIDI Ticks", "MIDI Bars"}, func(string) {})

			// turn into form items
			items := []*widget.FormItem{
				widget.NewFormItem("Max Notes Per Track", MaxNotesTXT),
				widget.NewFormItem("Length Type", LengthTXT),
			}

			// set default values
			MaxNotesTXT.SetText(app.Preferences().StringWithFallback("maxNotesPerTrack", "1000"))
			LengthTXT.SetSelected(app.Preferences().StringWithFallback("lengthType", "MIDI Ticks"))

			dialog.ShowForm("Settings", "Save", "Cancel", items, func(b bool) {
				if !b {
					return
				}

				// save values
				app.Preferences().SetString("maxNotesPerTrack", MaxNotesTXT.Text)
				app.Preferences().SetString("lengthType", LengthTXT.Selected)
			}, window)
		}),
	)

	// 1st row
	// hosts output file
	OutputTXT := widget.NewEntry()
	OutputLbl := widget.NewButton("Output", func() {
		fileDialog := dialog.NewFileSave(func(reader fyne.URIWriteCloser, _ error) {
			if reader != nil {
				p := reader.URI().Path()

				if path.Ext(p) != ".mid" {
					p += ".mid"
				}
				OutputTXT.SetText(p)
			}
		}, window)

		fileDialog.SetFilter(storage.NewExtensionFileFilter([]string{".mid"}))
		fileDialog.Show()
	})
	OutputLbl.Icon = theme.FileIcon()

	// 2nd row
	// hosts ppq, and bpm
	PPQLbl := createTxt("PPQ:")
	PPQTXT := widget.NewSelect([]string{"96", "192", "240", "480", "960", "1920", "3840", "8192"}, func(string) {})

	BPMLbl := createTxt("BPM:")
	BPMTXT := createNumberInput(0, 1000)

	// 3rd row
	// hosts ticks, NotesPerTrack
	TicksLbl := createTxt("MIDI Length:")
	TicksTXT := createNumberInput(0, -1)

	NotesLbl := createTxt("Notes:")
	NotesTXT := createNumberInput(0, -1)

	// 4th row
	// hosts min and max note length
	MinNoteLbl := createTxt("Min Note Length:")
	MinNoteTXT := createNumberInput(0, -1)

	MaxNoteLbl := createTxt("Max Note Length:")
	MaxNoteTXT := createNumberInput(0, -1)

	// output box
	OutputBox := widget.NewMultiLineEntry()
	OutputBox.SetText("Output will go here...")

	// create button
	createButton := widget.NewButton("Create", func() {
		var errors []string
		if err := OutputTXT.Validate(); err != nil {
			errors = append(errors, "output: "+err.Error())
		}
		if PPQTXT.Selected == "" {
			errors = append(errors, "ppq: cannot be empty")
		}
		if err := BPMTXT.Validate(); err != nil {
			errors = append(errors, "bpm: "+err.Error())
		}
		if err := TicksTXT.Validate(); err != nil {
			errors = append(errors, "ticks: "+err.Error())
		}
		if err := NotesTXT.Validate(); err != nil {
			errors = append(errors, "notes: "+err.Error())
		}

		if len(errors) > 0 {
			dialog.ShowInformation("Invalid Options", strings.Join(errors, "\n"), window)
		} else {
			OutputBox.SetText("")

			noteCount, err := strconv.Atoi(NotesTXT.Text)
			handleErr(err)
			ticks, err := strconv.Atoi(TicksTXT.Text)
			handleErr(err)
			minNoteLength, err := strconv.Atoi(MinNoteTXT.Text)
			handleErr(err)
			maxNoteLength, err := strconv.Atoi(MaxNoteTXT.Text)
			handleErr(err)
			maxNotesPerTrack, err := strconv.Atoi(app.Preferences().StringWithFallback("maxNotesPerTrack", "1000"))
			handleErr(err)
			ppq, err := strconv.Atoi(PPQTXT.Selected)
			handleErr(err)
			bpm, err := strconv.Atoi(BPMTXT.Text)
			handleErr(err)

			lengthType := app.Preferences().StringWithFallback("lengthType", "MIDI Ticks")

			if lengthType == "MIDI Bars" {
				ticks = int(float64(ticks) * float64(ppq) * 4)
			}

			OutputTXT.Disable()
			TicksTXT.Disable()
			MinNoteTXT.Disable()
			MaxNoteTXT.Disable()
			PPQTXT.Disable()
			BPMTXT.Disable()
			window.SetTitle("Random Note Generator (Running...)")

			OutputBox.SetText(OutputBox.Text + "creating tracks" + "\n")
			tracks := createTracks(
				noteCount,
				ticks,
				maxNoteLength,
				minNoteLength,
				maxNotesPerTrack,
				func(format string, args ...any) {
					OutputBox.SetText(OutputBox.Text + fmt.Sprintf(format, args...) + "\n")
				},
			)
			OutputBox.SetText(OutputBox.Text + "created tracks" + "\n")

			OutputBox.SetText(OutputBox.Text + "saving to midi" + "\n")
			createMIDI(OutputTXT.Text, ppq, bpm, tracks, func() {
				OutputBox.SetText(OutputBox.Text + "saved to midi" + "\n")

				OutputTXT.Enable()
				TicksTXT.Enable()
				MinNoteTXT.Enable()
				MaxNoteTXT.Enable()
				PPQTXT.Enable()
				BPMTXT.Enable()
				window.SetTitle("Random Note Generator")
			})
		}
	})

	// set default values
	OutputTXT.SetText(app.Preferences().StringWithFallback("outputPath", "output.mid"))

	PPQTXT.SetSelected(app.Preferences().StringWithFallback("ppq", "960"))
	BPMTXT.SetText(app.Preferences().StringWithFallback("bpm", "120"))

	TicksTXT.SetText(app.Preferences().StringWithFallback("ticks", "122880"))
	NotesTXT.SetText(app.Preferences().StringWithFallback("notes", "20000"))

	MinNoteTXT.SetText(app.Preferences().StringWithFallback("minNoteLength", "960"))
	MaxNoteTXT.SetText(app.Preferences().StringWithFallback("maxNoteLength", "1920"))

	// create content container
	content := container.NewBorder(
		container.NewVBox(
			container.New( // outputbtn outputtxt
				layout.NewFormLayout(),
				OutputLbl,
				OutputTXT,
			),
			container.New( // ppqlbl ppqtxt | bpmlbl bpmtxt
				layout.NewGridLayout(2),
				container.New(layout.NewFormLayout(), PPQLbl, PPQTXT),
				container.New(layout.NewFormLayout(), BPMLbl, BPMTXT),
			),
			container.New( // tickslbl tickstxt | noteslbl notestxt
				layout.NewGridLayout(2),
				container.New(layout.NewFormLayout(), TicksLbl, TicksTXT),
				container.New(layout.NewFormLayout(), NotesLbl, NotesTXT),
			),
			container.New( // minnotelbl minnotetxt | maxnotelbl maxnotetxt
				layout.NewGridLayout(2),
				container.New(layout.NewFormLayout(), MinNoteLbl, MinNoteTXT),
				container.New(layout.NewFormLayout(), MaxNoteLbl, MaxNoteTXT),
			),
			createButton,
		),
		HelpBar,
		nil,
		nil,
		container.New(
			layout.NewMaxLayout(),
			container.New(
				layout.NewMaxLayout(),
				OutputBox,
			),
		),
	)

	// set close intercept to save settings
	window.SetCloseIntercept(func() {
		logf("GUI closed, saving settings")

		app.Preferences().SetString("outputPath", OutputTXT.Text)
		app.Preferences().SetString("ppq", PPQTXT.Selected)
		app.Preferences().SetString("bpm", BPMTXT.Text)
		app.Preferences().SetString("ticks", TicksTXT.Text)
		app.Preferences().SetString("notes", NotesTXT.Text)

		window.Close()
	})

	// set content and show
	window.SetContent(content)
	window.ShowAndRun()
}

// helper functions to create objects with the same settings
func createTxt(text string) *canvas.Text {
	txt := canvas.NewText(text, color.White)
	txt.Alignment = fyne.TextAlignLeading
	txt.TextSize = 16

	return txt
}

func createNumberInput(min int, max int) *widget.Entry {
	entry := widget.NewEntry()
	entry.Validator = func(input string) error {
		if input == "" {
			return errors.New("cannot be empty")
		}

		num, err := strconv.Atoi(input)

		if err != nil {
			return errors.New("not a number")
		}

		if num < min {
			return errors.New("number too small")
		}

		if max != -1 && num > max {
			return errors.New("number too large")
		}

		return nil
	}
	return entry
}
