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
	"reflect"
	"sync"
	"time"
)

// Options struct for the simulation
type Options struct {
	TrackCircuitBased       bool           `json:"trackCircuitBased"`
	ClientToken             string         `json:"clientToken"`
	CurrentScore            int            `json:"currentScore"`
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

	simulation       *Simulation
	currentTime      Time
	currentTimeMutex sync.RWMutex
}

// ID func for options to that it implements SimObject. Returns an empty string.
func (o *Options) ID() string {
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
	if option == "CurrentTime" || option == "currentTime" {
		o.currentTimeMutex.Lock()
		defer o.currentTimeMutex.Unlock()
		t, ok := value.(Time)
		if !ok {
			return fmt.Errorf("cannot assign %v (%T) to currentTime (Time)", value, value)
		}
		o.currentTime = t
		return nil
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

// UnmarshalJSON for the options type
func (o *Options) UnmarshalJSON(data []byte) error {
	type optsUnmarshalType struct {
		TrackCircuitBased       bool           `json:"trackCircuitBased"`
		ClientToken             string         `json:"clientToken"`
		CurrentScore            int            `json:"currentScore"`
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
		CurrentTime             Time           `json:"CurrentTime"`
	}
	var auxOpts optsUnmarshalType
	if err := json.Unmarshal(data, &auxOpts); err!= nil {
		return err
	}
	o.TrackCircuitBased = auxOpts.TrackCircuitBased
	o.ClientToken = auxOpts.ClientToken
	o.CurrentScore = auxOpts.CurrentScore
	o.DefaultDelayAtEntry = auxOpts.DefaultDelayAtEntry
	o.DefaultMaxSpeed = auxOpts.DefaultMaxSpeed
	o.DefaultMinimumStopTime = auxOpts.DefaultMinimumStopTime
	o.DefaultSignalVisibility = auxOpts.DefaultSignalVisibility
	o.Description = auxOpts.Description
	o.TimeFactor = auxOpts.TimeFactor
	o.Title = auxOpts.Title
	o.Version = auxOpts.Version
	o.WarningSpeed = auxOpts.WarningSpeed
	o.WrongPlatformPenalty = auxOpts.WrongPlatformPenalty
	o.WrongDestinationPenalty = auxOpts.WrongDestinationPenalty
	o.LatePenalty = auxOpts.LatePenalty

	o.currentTimeMutex.Lock()
	defer o.currentTimeMutex.Unlock()
	o.currentTime = auxOpts.CurrentTime
	return nil
}

// CurrentTime returns the current time of the simulation
func (o *Options) CurrentTime() Time {
	o.currentTimeMutex.RLock()
	defer o.currentTimeMutex.RUnlock()
	return o.currentTime
}

// IncreaseTime increases the simulation time by the given step
func (o *Options) IncreaseTime(step time.Duration) {
	o.currentTimeMutex.Lock()
	defer o.currentTimeMutex.Unlock()
	o.currentTime = o.currentTime.Add(step)
}
