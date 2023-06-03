package model

import "fmt"

// Color is an RGB color displayed by the client.
type Color struct {
	Red   int
	Green int
	Blue  int
}

// Validate checks that the color RGB values are valid, returning an error if not.
func (c Color) Validate() error {
	if (c.Red >= 0 && c.Red <= 31) && (c.Green >= 0 && c.Green <= 31) && (c.Blue >= 0 && c.Blue <= 31) {
		return nil
	}

	return fmt.Errorf("rgb colors must be between 0 and 31")
}
