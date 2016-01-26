// +build js

package main

import "strconv"

type vector struct {
	yPos  int
	angle float64
	speed float64
}

func newVectorFromStrings(yPosS string, angleS string, speedS string) (*vector, error) {
	yPos, err := strconv.ParseInt(yPosS, 0, 32)
	if err != nil {
		return nil, err
	}
	angle, err := strconv.ParseFloat(angleS, 64)
	if err != nil {
		return nil, err
	}
	speed, err := strconv.ParseFloat(speedS, 64)
	if err != nil {
		return nil, err
	}
	return &vector{int(yPos), angle, speed}, nil
}
