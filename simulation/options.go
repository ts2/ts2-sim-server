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
	"fmt"
	"reflect"
)

// Options struct for the simulation
type Options struct {
	TrackCircuitBased       bool           `json:"trackCircuitBased"`
	ClientToken             string         `json:"clientToken"`
	CurrentScore            int            `json:"currentScore"`
	CurrentTime             Time           `json:"currentTime"`
	DefaultDelayAtEntry     DelayGenerator `json:"defaultDelayAtEntry"`
	DefaultMaxSpeed         float64        `json:"defaultMaxSpeed"`
	DefaultMinimumStopTime  DelayGenerator `json:"defaultMinimumStopTime"`
	DefaultSignalVisibility float64        `json:"defaultSignalVisibility"`
	Description             string         `json:"description"`
	TimeFactor              int            `json:"timeFactor"`
	Title                   string         `json:"title"`
	Version                 string         `json:"version"`
	WarningSpeed            float64        `json:"warningSpeed"`
	WrongPlatformPenalty    int            `json:"wrongPlatformPenalty"`
	WrongDestinationPenalty int            `json:"wrongDestinationPenalty"`
	LatePenalty             int            `json:"latePenalty"`

	simulation *Simulation
}

// ID func for options to that it implements SimObject. Returns an empty string.
func (o Options) ID() string {
	return ""
}

// Set the given option with the given value.
//
// option can be either the struct field name or the json key of the struct field.
func (o *Options) Set(option string, value interface{}) error {
	defer func() {
		o.simulation.sendEvent(&Event{Name: OptionsChangedEvent, Object: o})
	}()
	if value == nil {
		return fmt.Errorf("option %s cannot have nil value", option)
	}
	stVal := reflect.ValueOf(o).Elem()
	typ := stVal.Type()
	_, ok := typ.FieldByName(option)
	if ok {
		stVal.FieldByName(option).Set(reflect.ValueOf(value))
		return nil
	}
	for i := 0; i < typ.NumField(); i++ {
		if typ.Field(i).Tag.Get("json") == option {
			val := reflect.ValueOf(value)
			valTyp := val.Type()
			if valTyp.ConvertibleTo(typ.Field(i).Type) {
				val = val.Convert(typ.Field(i).Type)
				valTyp = val.Type()
			}
			if !valTyp.AssignableTo(typ.Field(i).Type) {
				return fmt.Errorf("cannot assign %v (%T) to %s (%s)", value, value, option, typ.Field(i).Type.Name())
			}
			stVal.Field(i).Set(val)
			return nil
		}
	}
	return fmt.Errorf("unknown option %s", option)
}
