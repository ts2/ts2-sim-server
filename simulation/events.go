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

const (
	// A ClockEvent is fired at each clock tick. Note that the actual time may not have changed.
	ClockEvent EventName = "clock"
	// A RouteActivatedEvent is emitted each time a route is successfully activated.
	RouteActivatedEvent EventName = "routeActivated"
	// A RouteDeactivatedEvent is emitted each time a route is successfully deactivated.
	RouteDeactivatedEvent EventName = "routeDeactivated"
	// TrainStoppedAtStationEvent is emitted each time a train arrives and stops at a scheduled station
	TrainStoppedAtStationEvent EventName = "trainStoppedAtStation"
	// TrainDepartedFromStationEvent is emitted each time a train departs from a station
	TrainDepartedFromStationEvent EventName = "trainDepartedFromStation"
	// SignalaspectChanged is emitted each time a Signal changes its aspect
	SignalaspectChanged EventName = "signalAspectChanged"
)

// Event is a wrapper around an object that is sent to the server hub to notify clients of a change.
type Event struct {
	Name   EventName
	Object interface{}
}
