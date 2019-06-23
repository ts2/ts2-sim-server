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

// An EventName is the name of a event
type EventName string

// Events that can be send to clients that add a listener to them.
const (
	ClockEvent                    EventName = "clock"
	StateChangedEvent             EventName = "stateChanged"
	OptionsChangedEvent           EventName = "optionsChanged"
	RouteActivatedEvent           EventName = "routeActivated"
	RouteDeactivatedEvent         EventName = "routeDeactivated"
	TrainStoppedAtStationEvent    EventName = "trainStoppedAtStation"
	TrainDepartedFromStationEvent EventName = "trainDepartedFromStation"
	TrainChangedEvent             EventName = "trainChanged"
	SignalaspectChangedEvent      EventName = "signalAspectChanged"
	TrackItemChangedEvent         EventName = "trackItemChanged"
	MessageReceivedEvent          EventName = "messageReceived"
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

// ID method to implement SimObject. Returns an empty string.
func (io IntObject) ID() string {
	return ""
}

// An BoolObject is a SimObject that wraps a single boolean value
type BoolObject struct {
	Value bool `json:"value"`
}

// ID method to implement SimObject. Returns an empty string.
func (bo BoolObject) ID() string {
	return ""
}
