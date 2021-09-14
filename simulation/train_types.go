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

import "encoding/json"

// TrainType defines a rolling stock type.
type TrainType struct {
	code         string
	Description  string   `json:"description"`
	EmergBraking float64  `json:"emergBraking"`
	Length       float64  `json:"length"`
	MaxSpeed     float64  `json:"maxSpeed"`
	StdAccel     float64  `json:"stdAccel"`
	StdBraking   float64  `json:"stdBraking"`
	ElementsStr  []string `json:"elements"`

	simulation *Simulation
}

// ID returns the unique identifier of this train type
func (tt *TrainType) ID() string {
	return tt.code
}

// setSimulation() attaches the simulation this TrainType is part of
func (tt *TrainType) setSimulation(sim *Simulation) {
	tt.simulation = sim
}

// initialize this train type
func (tt *TrainType) initialize(code string) {
	tt.code = code
}

// Elements returns the train types this TrainType is composed of.
func (tt *TrainType) Elements() []*TrainType {
	res := make([]*TrainType, 0)
	for _, code := range tt.ElementsStr {
		res = append(res, tt.simulation.TrainTypes[code])
	}
	return res
}

// MarshalJSON for the TrainType type
func (tt *TrainType) MarshalJSON() ([]byte, error) {
	type auxTT struct {
		ID           string   `json:"id"`
		Description  string   `json:"description"`
		EmergBraking float64  `json:"emergBraking"`
		Length       float64  `json:"length"`
		MaxSpeed     float64  `json:"maxSpeed"`
		StdAccel     float64  `json:"stdAccel"`
		StdBraking   float64  `json:"stdBraking"`
		ElementsStr  []string `json:"elements"`
	}
	att := auxTT{
		ID:           tt.ID(),
		Description:  tt.Description,
		EmergBraking: tt.EmergBraking,
		Length:       tt.Length,
		MaxSpeed:     tt.MaxSpeed,
		StdAccel:     tt.StdAccel,
		StdBraking:   tt.StdBraking,
		ElementsStr:  tt.ElementsStr,
	}
	return json.Marshal(att)
}
