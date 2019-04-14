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

import "fmt"

// An EventName is the name of a event
type EventName string

// Events that can be send to clients that add a listener to them.
const (
	ClockEvent                    EventName = "clock"
	RouteActivatedEvent           EventName = "routeActivated"
	RouteDeactivatedEvent         EventName = "routeDeactivated"
	TrainStoppedAtStationEvent    EventName = "trainStoppedAtStation"
	TrainDepartedFromStationEvent EventName = "trainDepartedFromStation"
	TrainChanged                  EventName = "trainChanged"
	SignalaspectChanged           EventName = "signalAspectChanged"
	TrackItemChanged              EventName = "trackItemChanged"
	MessageReceived               EventName = "messageReceived"
	ScoreChanged                  EventName = "scoreChanged"
)

// A SimObject can be serialized in an event
type SimObject interface {
	ID() string
}

// Event is a wrapper around an object that is sent to the server hub to notify clients of a change.
type Event struct {
	Name   EventName
	Object SimObject
}

// An IntObject is a SimObject that wraps a single integer value
type IntObject struct {
	Value int `json:"value"`
}

// ID method to implement SimObject. Returns the Value as a string.
func (io IntObject) ID() string {
	return fmt.Sprint(io.Value)
}
