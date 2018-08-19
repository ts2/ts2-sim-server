// Copyright (C) 2008-2018 by Nicolas Piganeau and the TS2 TEAM
// (See AUTHORS file)
//
// This program is free software; you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation; either version 2 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program; if not, write to the
// Free Software Foundation, Inc.,
// 59 Temple Place - Suite 330, Boston, MA  02111-1307, USA.

package simulation

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

type delayTuplet [3]int

// DelayGenerator is a probability distribution for a duration in seconds
// and is used to generate random delays for trains.
//
//  - The `data` field is a list of tuplets (an array of 3 integers).
//  - Each tuple defines in order:
//	   - A lower bound
//	   - An upper bound
//	   - A probability in percent of the value to be inside the defined bounds.
//
//	e.g. [[0 100 80] [100 500 20]] means that when a value will be yielded by
//	this DelayGenerator, it will have 80% chance of being between 0 and 100, and
//	20% chance of being between 100 and 500.
type DelayGenerator struct {
	data []delayTuplet
}

func (dg *DelayGenerator) UnmarshalJSON(data []byte) error {
	var field []delayTuplet
	if err := json.Unmarshal(data, &field); err != nil {
		// Failed with delayTuplet[], so try a single value eg 0
		var single int
		if err := json.Unmarshal(data, &single); err != nil {
			return fmt.Errorf("DelayGenerator.UnmarshalJSON(): Unparsable JSON: %s", data)
		}
		dg.data = []delayTuplet{{single, single, 100}}

	} else {
		dg.data = field
	}
	return nil
}

// Time type for the simulation (HH:MM:SS).
//
// Valid Time objects start on 0000-01-02.
type Time struct {
	sync.RWMutex
	time.Time
}

// UnmarshalJSON for the Time type
func (h *Time) UnmarshalJSON(data []byte) error {
	var hourStr string
	if err := json.Unmarshal(data, &hourStr); err != nil {
		return fmt.Errorf("times should be encoded as 00:00:00 strings in JSON, got %s instead", data)
	}
	*h = ParseTime(hourStr)
	return nil
}

// MarshalJSON for the Time type
func (h Time) MarshalJSON() ([]byte, error) {
	return json.Marshal(h.Time.Format("15:04:05"))
}

// ParseTime returns a Time object from its string representation in format 15:04:05
func ParseTime(data string) Time {
	t, err := time.Parse("15:04:05", data)
	if err != nil {
		return Time{}
	}
	// We add 24 hours to make a difference between 00:00:00 and an empty Time
	return Time{
		Time: t.Add(24 * time.Hour),
	}
}

// Add returns the time h + duration .
func (h Time) Add(duration time.Duration) Time {
	newTime := h
	newTime.Time = h.Time.Add(duration)
	return newTime
}

// Sub returns the duration t-u. If the result exceeds the maximum (or minimum)
// value that can be stored in a Duration, the maximum (or minimum) duration
// will be returned.
// To compute t-d for a duration d, use t.Add(-d).
func (h Time) Sub(u Time) time.Duration {
	return h.Time.Sub(u.Time)
}

var _ json.Marshaler = Time{}
var _ json.Unmarshaler = new(Time)
