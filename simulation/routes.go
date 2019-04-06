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
)

// A RoutesManager checks if a route is activable or deactivable.
type RoutesManager interface {
	// Name returns a description of this routesManager that can be displayed
	// to the user if one of this managers method returns false.
	Name() string
	// CanActivate returns an error if the given route cannot be Activated
	CanActivate(r *Route) error
	// CanDeactivate returns an error if the given route cannot be Deactivated
	CanDeactivate(r *Route) error
}

// RouteState represents the state of a Route at a given time and instance
type RouteState uint8

const (
	// Deactivated =  The route is not active
	Deactivated RouteState = 0

	// Activated =  The route is active but will be destroyed by the first train using it
	Activated RouteState = 1

	// Persistent =  The route is set and will remain after train passage
	Persistent RouteState = 2
)

// A Route is a path between two signals.
//
// If a route is Activated, the path is selected, and the signals at the beginning
// and the end of the route are changed and the conflicting possible other routes
// are inhibited. Routes are static and defined in the game file. The player can
// only activate or deactivate them.
type Route struct {
	routeID       string                    `json:"id"`
	BeginSignalId string                    `json:"beginSignal"`
	EndSignalId   string                    `json:"endSignal"`
	InitialState  RouteState                `json:"initialState"`
	Directions    map[string]PointDirection `json:"directions"`
	State         RouteState                `json:"state"`
	Positions     []Position                `json:"-"`

	simulation *Simulation
	triggers   []func(*Route)
}

// ID returns the unique identifier of this route
func (r *Route) ID() string {
	return r.routeID
}

// BeginSignal returns the SignalItem at which this Route starts.
func (r *Route) BeginSignal() *SignalItem {
	return r.simulation.TrackItems[r.BeginSignalId].(*SignalItem)
}

// EndSignal returns the SignalItem at which this Route ends.
func (r *Route) EndSignal() *SignalItem {
	return r.simulation.TrackItems[r.EndSignalId].(*SignalItem)
}

// Equals returns true if this Route is the same as other, that is they
// have the same routeID.
func (r *Route) Equals(other *Route) bool {
	return r.routeID == other.routeID
}

// IsActive returns true if this Route is active
func (r *Route) IsActive() bool {
	return r.State == Activated || r.State == Persistent
}

// addTrigger adds the given function to the list of function that will be
// called when this Route is activated or deactivated.
func (r *Route) addTrigger(trigger func(*Route)) {
	r.triggers = append(r.triggers, trigger)
}

// Activate the given route. If the route cannot be Activated, an error is returned.
func (r *Route) Activate(persistent bool) error {
	for _, rm := range routesManagers {
		if err := rm.CanActivate(r); err != nil {
			return fmt.Errorf("%s vetoed route activation: %s", rm.Name(), err)
		}
	}
	for _, pos := range r.Positions {
		pos.TrackItem().setActiveRoute(r, pos.PreviousItem())
	}
	r.EndSignal().previousActiveRoute = r
	r.BeginSignal().nextActiveRoute = r
	r.State = Activated
	if persistent {
		r.State = Persistent
	}
	for _, t := range r.triggers {
		t(r)
	}
	r.simulation.sendEvent(&Event{
		Name:   RouteActivatedEvent,
		Object: r,
	})
	r.BeginSignal().updateSignalState()
	return nil
}

// Deactivate the given route. If the route cannot be Deactivated, an error is returned.
func (r *Route) Deactivate() error {
	for _, rm := range routesManagers {
		if rm.CanDeactivate(r) != nil {
			return fmt.Errorf("%s vetoed route deactivation", rm.Name())
		}
	}
	r.BeginSignal().resetNextActiveRoute(r)
	r.EndSignal().resetPreviousActiveRoute(nil)
	for _, pos := range r.Positions {
		if pos.TrackItem().ActiveRoute() != nil && pos.TrackItem().ActiveRoute().routeID != r.routeID {
			continue
		}
		pos.TrackItem().setActiveRoute(nil, nil)
	}
	r.State = Deactivated
	for _, t := range r.triggers {
		t(r)
	}
	r.simulation.sendEvent(&Event{
		Name:   RouteDeactivatedEvent,
		Object: r,
	})
	r.BeginSignal().updateSignalState()
	return nil
}

// setSimulation sets the Simulation this Route is part of.
func (r *Route) setSimulation(sim *Simulation) {
	r.simulation = sim
}

// initialize does initial steps necessary to use this route
func (r *Route) initialize(routeNum string) error {
	// Set route routeID
	r.routeID = routeNum

	// Initialize state to initial state
	r.State = r.InitialState

	// Populate Positions slice
	pos := Position{
		TrackItemID:    r.BeginSignal().ID(),
		PreviousItemID: r.BeginSignal().PreviousItem().ID(),
		PositionOnTI:   0,
		simulation:     r.simulation}
	for !pos.IsOut() {
		r.Positions = append(r.Positions, pos)
		if pos.TrackItem().ID() == r.EndSignal().ID() {
			return nil
		}
		dir := DirectionCurrent
		if pi, ok := pos.TrackItem().(*PointsItem); ok {
			dir, ok = r.Directions[pi.ID()]
			if !ok {
				switch pos.PreviousItemID {
				case pi.ReverseTiId:
					dir = DirectionReversed
				case pi.PreviousTiID, pi.NextTiID:
					dir = DirectionNormal
				default:
					return fmt.Errorf("route Error: unable to find direction for points %s", pi.ID())
				}
				r.Directions[pi.ID()] = dir
			}
		}

		pos = pos.Next(dir)
	}

	return fmt.Errorf("route Error: unable to link signal %s to signal %s", r.BeginSignalId, r.EndSignalId)
}

// UnmarshalJSON for the Route type
func (r *Route) UnmarshalJSON(data []byte) error {
	type auxRoute struct {
		BeginSignalId string                    `json:"beginSignal"`
		EndSignalId   string                    `json:"endSignal"`
		InitialState  RouteState                `json:"initialState"`
		Directions    map[string]PointDirection `json:"directions"`
	}
	var rawRoute auxRoute
	if err := json.Unmarshal(data, &rawRoute); err != nil {
		return fmt.Errorf("unable to decode simulation JSON: %s", err)
	}
	r.BeginSignalId = rawRoute.BeginSignalId
	r.EndSignalId = rawRoute.EndSignalId
	r.InitialState = rawRoute.InitialState
	r.Directions = make(map[string]PointDirection)
	for tiID, dir := range rawRoute.Directions {
		r.Directions[tiID] = dir
	}
	return nil
}

// MarshalJSON for the Route type
func (r *Route) MarshalJSON() ([]byte, error) {
	type auxRoute struct {
		ID            string                    `json:"id"`
		BeginSignalId string                    `json:"beginSignal"`
		EndSignalId   string                    `json:"endSignal"`
		InitialState  RouteState                `json:"initialState"`
		Directions    map[string]PointDirection `json:"directions"`
		State         RouteState                `json:"state"`
	}
	ar := auxRoute{
		ID:            r.ID(),
		BeginSignalId: r.BeginSignalId,
		EndSignalId:   r.EndSignalId,
		InitialState:  r.InitialState,
		Directions:    r.Directions,
		State:         r.State,
	}
	d, err := json.Marshal(ar)
	return d, err
}
