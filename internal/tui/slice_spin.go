package tui

import "github.com/charmbracelet/bubbles/spinner"

type namedSpinner struct {
	name string
	spin spinner.Model
}

type namedSpinners []namedSpinner

func (s namedSpinners) remove(name string) namedSpinners {
	for i, p := range s {
		if p.name == name {
			return append(s[:i], s[i+1:]...)
		}
	}

	return s
}

func (s namedSpinners) lastSpinner() spinner.Model {
	return s[len(s)-1].spin
}
