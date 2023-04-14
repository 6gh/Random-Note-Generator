# Random Note Generator
A simple tool to randomly generate notes

![Preview of the GUI](https://i.imgur.com/Mlp5tnZ.png)

## Purpose

This project is to quickly randomly generate a MIDI file with a specified number of notes. In [Black MIDI](https://en.wikipedia.org/wiki/Black_Midi) there are many times when you need random notes to reach an exact note count. This aims to help out in that.

This program will generate a random number of notes and split them into separate tracks, that you specify. Thus, this might require you to do a bit of tinkering to get the perfect result.

## Usage

Download the [latest release](https://github.com/6gh/Empty-Track-Creator/releases/latest). Currently, the only built release is for windows. This is due to me not having a Linux or Mac machine, so I am not able to verify that it works on these OSes.

Open the program and set your settings:
- Output - The output path to your MIDI. Must end in .mid and can be relative pathing
- PPQ - The PPQ of the output MIDI
- BPM - The BPM of the output MIDI
- MIDI Length - How long the MIDI can be, in ticks or bars (see below). All notes will be cut off at the max length, if trim notes is true (see below)
- Notes - The amount of notes you want to generate
- Min Note Length - The shortest a random note can be in ticks
- Max Note Length - The longest a random note can be in ticks

Click the cog at the bottom to set additional settings:
- Max Notes Per Track - The number of notes that a single track can contain, before creating a new one
- Length Type - Whether the `MIDI Length` should be in Ticks or Bars. If it is in ticks, the length will be dependent on the PPQ, and you will have to calculate it yourself. If it is in bars, the length will be translated to ticks for you
- Trim Notes - Whether or not to trim the notes which go beyond the MIDI length
- Note Velocity - Changes the notes' velocities
- Note Channel - Changes what channel the notes will be generated in

## Building 

You will need to install the packages required using Go and also follow [Fyne getting started guide](https://developer.fyne.io/started/) to install and use fyne (gui framework). After that just use `fyne package` and you will get your executable.

## License

[MIT](https://github.com/6gh/Empty-Track-Creator/blob/master/LICENSE)
