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

type TrainStatus uint8

const (
	// Not yet entered on the scene
	inactive TrainStatus = 0

	// Running with a positive speed
	running TrainStatus = 10

	// Scheduled stop, e.g. at a station
	stopped TrainStatus = 20

	// Unscheduled stop, e.g. at a red signal
	waiting TrainStatus = 30

	// Exited the area
	out TrainStatus = 40

	// Ended its service and no new service assigned
	endOfService TrainStatus = 50
)

// Train is a stock of `TrainType` running on a track at a certain speed and to which
// is assigned a `Service`.
type Train struct {
	AppearTime     Time           `json:"appearTime"`
	InitialDelay   DelayGenerator `json:"initialDelay"`
	InitialSpeed   float64        `json:"initialSpeed"`
	NextPlaceIndex int            `json:"nextPlaceIndex"`
	ServiceCode    string         `json:"serviceCode"`
	Speed          float64        `json:"speed"`
	Status         TrainStatus    `json:"status"`
	StoppedTime    int            `json:"stoppedTime"`
	TrainHeadRepr  PositionRepr   `json:"trainHead"`
	TrainTypeCode  string         `json:"trainTypeCode"`
	TrainHead      *Position

	simulation *Simulation
}

// setSimulation attaches the Simulation to this Train, and in is part of and initialized TrainHead.
func (t *Train) setSimulation(sim *Simulation) {
	t.simulation = sim
	if th, err := NewPosition(sim, t.TrainHeadRepr); err != nil {
		panic(err)
	} else {
		t.TrainHead = th
	}
}

// Service returns a pointer to the Service assigned to this Train, or nil if no
// Service is assigned.
func (t *Train) Service() *Service {
	return t.simulation.Services[t.ServiceCode]
}

// TrainType returns a pointer to the TrainType that this Train is running.
func (t *Train) TrainType() *TrainType {
	return t.simulation.TrainTypes[t.TrainTypeCode]
}
