package main

import (
	"math/rand"
	"os"
	"sort"

	"gitlab.com/gomidi/midi/v2"
	"gitlab.com/gomidi/midi/v2/smf"
)

// Creates an array of tracks
func createTracks(noteCount int, ticks int, maxNoteLength int, minNoteLength int, maxNotesPerTrack int, trimNotes bool, minVelocity int, maxVelocity int, noteChannel string, logger func(format string, a ...any)) []smf.Track {
	var (
		tracks               []smf.Track
		remainingNotes       = noteCount
		specifiedChannel     = -1
		currentChannelNumber = 0
		trackCount           = 0
	)

	// get channel
	// if noteChannel is "All (Skip Drums)", set channel to -1
	// if noteChannel is "All", set channel to -2
	switch noteChannel {
	case "All (Skip Drums)":
		specifiedChannel = -1
	case "All":
		specifiedChannel = -2
	case "1":
		specifiedChannel = 0
	case "2":
		specifiedChannel = 1
	case "3":
		specifiedChannel = 2
	case "4":
		specifiedChannel = 3
	case "5":
		specifiedChannel = 4
	case "6":
		specifiedChannel = 5
	case "7":
		specifiedChannel = 6
	case "8":
		specifiedChannel = 7
	case "9":
		specifiedChannel = 8
	case "10 (Drums)":
		specifiedChannel = 9
	case "11":
		specifiedChannel = 10
	case "12":
		specifiedChannel = 11
	case "13":
		specifiedChannel = 12
	case "14":
		specifiedChannel = 13
	case "15":
		specifiedChannel = 14
	case "16":
		specifiedChannel = 15
	}

	logger("generating notes")
	for i := 0; i < noteCount; {
		if specifiedChannel == -1 {
			// user selected "All (Skip Drums)"
			// set the current channel based the current track number
			// if the current channel is 9 (drums), skip it
			currentChannelNumber = trackCount % 16
			if currentChannelNumber == 9 {
				currentChannelNumber++ // skip drums
				trackCount++           // increment track count to avoid double ch 11
			}
		} else if specifiedChannel == -2 {
			// user selected "All"
			// set the current channel based the current track number
			currentChannelNumber = trackCount % 16
		} else {
			// user selected a specific channel
			currentChannelNumber = specifiedChannel
		}

		// calculate the number of notes to add to the track
		var nc int
		if remainingNotes > maxNotesPerTrack {
			// if there are more notes left than the max notes per track, set the number of notes to the max notes per track
			// this generates a track with the max notes per track
			nc = maxNotesPerTrack
			remainingNotes = remainingNotes - maxNotesPerTrack
			i = i + maxNotesPerTrack
		} else {
			// if there are less notes left, or equal to, the max notes per track, set the remaining notes to 0
			// this generates a track will all the notes left
			nc = remainingNotes
			remainingNotes = 0
			i = i + noteCount
		}

		logger("generating track (ch %d) with %d notes | notes left: %d", currentChannelNumber+1, nc, remainingNotes)

		track := createTrack(nc, ticks, maxNoteLength, minNoteLength, trimNotes, minVelocity, maxVelocity, uint8(currentChannelNumber))
		tracks = append(tracks, track)
		trackCount++
	}

	logger("generated %d tracks", len(tracks))
	return tracks
}

// Creates a track, with a specified number of notes
func createTrack(noteCount int, ticks int, maxNoteLength int, minNoteLength int, trimNotes bool, minVelocity int, maxVelocity int, channel uint8) smf.Track {
	var (
		track  smf.Track
		events []NoteEvent
	)

	// create notes
	for i := 0; i < noteCount; i++ {
		noteStart := rand.Intn(ticks)                                          // get a random start time between 0 and the length of the midi
		noteDuration := rand.Intn(maxNoteLength-minNoteLength) + minNoteLength // get a random duration between min length and the max length of a note
		noteKey := uint8(rand.Intn(128))                                       // get a random key between 0 and 127 (C0 - G10)
		noteEnd := noteStart + noteDuration                                    // calculate the end time
		if trimNotes && noteEnd > ticks {                                      // only cut notes if cutNotes is true
			noteEnd = ticks // if end time is greater than the length of the midi, set it to the length of the midi
		}

		// add note event
		// notes = append(notes, Notes{uint32(noteStart), uint32(noteEnd), noteKey})
		events = append(events, NoteEvent{uint32(noteStart), noteKey, true})
		events = append(events, NoteEvent{uint32(noteEnd), noteKey, false})
	}

	// sort notes by start time
	sort.Sort(EventSorter(events))

	// iterate through notes again
	for i := 0; i < len(events); i++ {
		// this is done because the midi library uses a relative tick system to add events
		// (ticks start from the previous event's end tick)
		// so we need to calculate the difference between the current note start and the previous note end
		event := events[i] // get the current note

		var tick uint32
		if i > 0 { // if this is not the first note
			prevNote := events[i-1] // get the previous note
			tick = event.tick - prevNote.tick
		} else { // if this is the first note
			tick = event.tick
		}

		if event.noteOn { // add note on event
			var noteVelocity int
			if minVelocity == maxVelocity { // if min and max velocity are the same, set the velocity to that
				noteVelocity = minVelocity
			} else {
				noteVelocity = rand.Intn(maxVelocity-minVelocity) + minVelocity // get a random velocity between min and max
			}
			track.Add(tick, midi.NoteOn(channel, event.key, uint8(noteVelocity)))
		} else { // add note off event
			track.Add(tick, midi.NoteOff(channel, event.key))
		}
	}
	track.Close(0)
	return track
}

// Creates a midi file, adding the tracks given
func createMIDI(midiPath string, ppq int, bpm int, tracks []smf.Track, callback func()) {
	// create vars
	var (
		resolution = smf.MetricTicks(ppq)
		firstTrack smf.Track
		midiData   = smf.New()
	)

	// set midi data
	// ppq, meta track
	midiData.TimeFormat = resolution                 // set ppq
	firstTrack.Add(0, smf.MetaTrackSequenceName("")) // add a blank track name
	firstTrack.Add(0, smf.MetaTempo(float64(bpm)))   // set bpm
	firstTrack.Close(0)
	midiData.Add(firstTrack)

	// add all tracks provided
	for i := 0; i < len(tracks); i++ {
		midiData.Add(tracks[i])
	}

	// open, or create, the midi file
	file, err := os.OpenFile(midiPath, os.O_CREATE|os.O_WRONLY, 0644)
	handleErr(err)

	// write the midi data to the file
	_, err = midiData.WriteTo(file)
	handleErr(err)

	// close the file
	err = file.Close()
	handleErr(err)

	callback() // call the callback function
}

type NoteEvent struct {
	tick   uint32
	key    uint8
	noteOn bool
}

type EventSorter []NoteEvent

func (a EventSorter) Len() int           { return len(a) }
func (a EventSorter) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a EventSorter) Less(i, j int) bool { return a[i].tick < a[j].tick }
