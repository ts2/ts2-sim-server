// Copyright (C) 2008-2019 by Nicolas Piganeau and the TS2 TEAM
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
	"strings"
)

// nextActiveRoute is true if a route starting from this Signal is active
type NextActiveRoute struct{}

// Code of the ConditionType, uniquely defines this ConditionType
func (nar NextActiveRoute) Code() string {
	return "NEXT_ROUTE_ACTIVE"
}

// Solve returns if the condition is met for the given SignalItem and parameters
func (nar NextActiveRoute) Solve(item *SignalItem, values []string, params []string) bool {
	return item.nextActiveRoute != nil && item.nextActiveRoute.IsActive()
}

// SetupTriggers installs needed triggers for the given SignalItem, with the
// given Condition.
func (nar NextActiveRoute) SetupTriggers(item *SignalItem, params []string) {}

// ---------------------------------------------------------------------------------------------------------------

// previousActiveRoute is true if a route ending at this Signal is active
type PreviousActiveRoute struct{}

// Code of the ConditionType, uniquely defines this ConditionType
func (par PreviousActiveRoute) Code() string {
	return "PREVIOUS_ROUTE_ACTIVE"
}

// Solve returns if the condition is met for the given SignalItem and parameters
func (par PreviousActiveRoute) Solve(item *SignalItem, values []string, params []string) bool {
	return item.previousActiveRoute != nil && (item.previousActiveRoute.IsActive() || item.previousActiveRoute.IsDestroying())
}

// SetupTriggers installs needed triggers for the given SignalItem, with the
// given Condition.
func (par PreviousActiveRoute) SetupTriggers(item *SignalItem, params []string) {}

// ---------------------------------------------------------------------------------------------------------------

// RouteSetAcross is true if a route is active across this signal, in the same direction
// but neither starting nor ending at this signal.
type RouteSetAcross struct{}

// Code of the ConditionType, uniquely defines this ConditionType
func (rsa RouteSetAcross) Code() string {
	return "ROUTE_SET_ACROSS"
}

// Solve returns if the condition is met for the given SignalItem and parameters
func (rsa RouteSetAcross) Solve(item *SignalItem, values []string, params []string) bool {
	if item.ActiveRoute() != nil {
		positions := item.ActiveRoute().Positions
		for _, pos := range positions[1 : len(positions)-1] {
			if item.IsOnPosition(pos) {
				return true
			}
		}
	}
	return false
}

// SetupTriggers installs needed triggers for the given SignalItem, with the
// given Condition.
func (rsa RouteSetAcross) SetupTriggers(item *SignalItem, params []string) {}

// ---------------------------------------------------------------------------------------------------------------

// TrainNotPresentOnNextRoute is true if there is no train ahead of this signal and
// before the end of the next active route. If no route is set, the condition is always false.
type TrainNotPresentOnNextRoute struct{}

// Code of the ConditionType, uniquely defines this ConditionType
func (tnpnr TrainNotPresentOnNextRoute) Code() string {
	return "TRAIN_NOT_PRESENT_ON_NEXT_ROUTE"
}

// Solve returns if the condition is met for the given SignalItem and parameters
func (tnpnr TrainNotPresentOnNextRoute) Solve(item *SignalItem, values []string, params []string) bool {
	if item.nextActiveRoute == nil {
		return false
	}
	for _, pos := range item.nextActiveRoute.Positions {
		if pos.TrackItem().TrainPresent() {
			return false
		}
	}
	return true
}

// SetupTriggers installs needed triggers for the given SignalItem, with the
// given Condition.
func (tnpnr TrainNotPresentOnNextRoute) SetupTriggers(item *SignalItem, params []string) {}

// ---------------------------------------------------------------------------------------------------------------

// TrainNotPresentBeforeNextSignal is true if there is no train ahead of this signal and
// before the next signal on the line.
// Signal aspects ending with ! can be added in the list to discard the given signal and look
// up to the next one.
type TrainNotPresentBeforeNextSignal struct{}

// Code of the ConditionType, uniquely defines this ConditionType
func (tnpbns TrainNotPresentBeforeNextSignal) Code() string {
	return "TRAIN_NOT_PRESENT_BEFORE_NEXT_SIGNAL"
}

// Solve returns if the condition is met for the given SignalItem and parameters
func (tnpbns TrainNotPresentBeforeNextSignal) Solve(item *SignalItem, values []string, params []string) bool {
mainLoop:
	for cur := item.Position(); !cur.IsOut(); cur = cur.Next(DirectionCurrent) {
		if cur.TrackItem().TrainPresent() {
			return false
		}
		if !cur.Equals(item.Position()) && cur.TrackItem().Type() == TypeSignal && cur.TrackItem().IsOnPosition(cur) {
			for _, v := range values {
				if !strings.HasSuffix(v, "!") {
					// Ignore values not ending with !
					continue
				}
				aspectName := strings.TrimSuffix(v, "!")
				if cur.TrackItem().(*SignalItem).ActiveAspect().Name == aspectName {
					continue mainLoop
				}
			}
			break
		}
	}
	return true
}

// SetupTriggers installs needed triggers for the given SignalItem, with the
// given Condition.
func (tnpbns TrainNotPresentBeforeNextSignal) SetupTriggers(item *SignalItem, params []string) {}

// ---------------------------------------------------------------------------------------------------------------

// TrainNotPresentOnItems is true if there is no train on the track items defined by params.
type TrainNotPresentOnItems struct{}

// Code of the ConditionType, uniquely defines this ConditionType
func (tnpoi TrainNotPresentOnItems) Code() string {
	return "TRAIN_NOT_PRESENT_ON_ITEMS"
}

// Solve returns if the condition is met for the given SignalItem and parameters
func (tnpoi TrainNotPresentOnItems) Solve(item *SignalItem, values []string, params []string) bool {
	for _, id := range params {
		if item.Simulation().TrackItems[id].TrainPresent() {
			return false
		}
	}
	return true
}

// SetupTriggers installs needed triggers for the given SignalItem, with the
// given Condition.
func (tnpoi TrainNotPresentOnItems) SetupTriggers(item *SignalItem, params []string) {
	for _, id := range params {
		ti, ok := item.Simulation().TrackItems[id]
		if !ok {
			panic(fmt.Errorf("TrainNotPresentOnItems: error in simulation definition.\n"+
				"SignalItem %s reference unknown TrackItem %s", item.ID(), id))
		}
		ti.addTrigger(func(t TrackItem) {
			item.updateSignalState()
		})
	}

}

// ---------------------------------------------------------------------------------------------------------------

// TrainPresentOnItems is true if there a train on all the track items defined by custom property.
type TrainPresentOnItems struct{}

// Code of the ConditionType, uniquely defines this ConditionType
func (tpoi TrainPresentOnItems) Code() string {
	return "TRAIN_PRESENT_ON_ITEMS"
}

// Solve returns if the condition is met for the given SignalItem and parameters
func (tpoi TrainPresentOnItems) Solve(item *SignalItem, values []string, params []string) bool {
	for _, id := range params {
		if !item.Simulation().TrackItems[id].TrainPresent() {
			return false
		}
	}
	return true
}

// SetupTriggers installs needed triggers for the given SignalItem, with the
// given Condition.
func (tpoi TrainPresentOnItems) SetupTriggers(item *SignalItem, params []string) {
	for _, id := range params {
		ti, ok := item.Simulation().TrackItems[id]
		if !ok {
			panic(fmt.Errorf("TrainPresentOnItems: error in simulation definition.\n"+
				"SignalItem %s reference unknown TrackItem %s", item.ID(), id))
		}
		ti.addTrigger(func(t TrackItem) {
			item.updateSignalState()
		})
	}
}

// ---------------------------------------------------------------------------------------------------------------

// RouteSet is true if at least one of the routes, the id of which is defined by custom property is active.
// These routes don't have to start at this signal.
type RouteSet struct{}

// Code of the ConditionType, uniquely defines this ConditionType
func (rs RouteSet) Code() string {
	return "ROUTES_SET"
}

// Solve returns if the condition is met for the given SignalItem and parameters
func (rs RouteSet) Solve(item *SignalItem, values []string, params []string) bool {
	for _, id := range params {
		if item.Simulation().Routes[id].IsActive() {
			return true
		}
	}
	return false
}

// SetupTriggers installs needed triggers for the given SignalItem, with the
// given Condition.
func (rs RouteSet) SetupTriggers(item *SignalItem, params []string) {
	for _, id := range params {
		r, ok := item.Simulation().Routes[id]
		if !ok {
			panic(fmt.Errorf("RouteSet: error in simulation definition.\n"+
				"SignalItem %s reference unknown Route %s", item.ID(), id))
		}
		r.addTrigger(func(r *Route) {
			item.updateSignalState()
		})
	}
}

// ---------------------------------------------------------------------------------------------------------------

// NextSignalAspects is true if the next signal is showing one of the aspects given.
//
// If one of the aspect names finishes with a '!' and the next signal aspect matches this aspect,
// then the next signal aspect is ignored and the aspect further on the line is checked with the same data.
type NextSignalAspects struct{}

// Code of the ConditionType, uniquely defines this ConditionType
func (nsa NextSignalAspects) Code() string {
	return "NEXT_SIGNAL_ASPECTS"
}

// checkSignalAspect checks if the given signal has one of the given aspect name.
func checkSignalAspect(signal *SignalItem, aspectNames []string, previous ...bool) bool {
	if len(previous) > 100 {
		// Prevent infinite recursion
		return false
	}
	if signal != nil {
		for _, v := range aspectNames {
			if strings.HasSuffix(v, "!") {
				aspectName := strings.TrimSuffix(v, "!")
				if aspectName == signal.ActiveAspect().Name {
					return checkSignalAspect(signal.getNextSignal(), aspectNames, append(previous, true)...)
				}
				continue
			}
			if v == signal.ActiveAspect().Name {
				return true
			}
		}
	}
	return false
}

// Solve returns if the condition is met for the given SignalItem and parameters
func (nsa NextSignalAspects) Solve(item *SignalItem, values []string, params []string) bool {
	return checkSignalAspect(item.getNextSignal(), values)
}

// SetupTriggers installs needed triggers for the given SignalItem, with the
// given Condition.
func (nsa NextSignalAspects) SetupTriggers(item *SignalItem, params []string) {}

// ---------------------------------------------------------------------------------------------------------------

// RouteExitSignalAspects is true if the exit signal of the route starting at this signal is
// showing one of the aspects given.
// If no route is set from this signal, the condition is always false.
type RouteExitSignalAspects struct{}

// Code of the ConditionType, uniquely defines this ConditionType
func (resa RouteExitSignalAspects) Code() string {
	return "ROUTE_EXIT_SIGNAL_ASPECTS"
}

// Solve returns if the condition is met for the given SignalItem and parameters
func (resa RouteExitSignalAspects) Solve(item *SignalItem, values []string, params []string) bool {
	if item.nextActiveRoute == nil {
		return false
	}
	nextSignal := item.nextActiveRoute.EndSignal()
	if nextSignal != nil {
		for _, v := range values {
			if v == nextSignal.ActiveAspect().Name {
				return true
			}
		}
	}
	return false
}

// SetupTriggers installs needed triggers for the given SignalItem, with the
// given Condition.
func (resa RouteExitSignalAspects) SetupTriggers(item *SignalItem, params []string) {}

// ---------------------------------------------------------------------------------------------------------------

func init() {
	signalConditionTypes = make(map[string]ConditionType)
	nar := NextActiveRoute{}
	signalConditionTypes[nar.Code()] = nar
	par := PreviousActiveRoute{}
	signalConditionTypes[par.Code()] = par
	rsa := RouteSetAcross{}
	signalConditionTypes[rsa.Code()] = rsa
	tnponr := TrainNotPresentOnNextRoute{}
	signalConditionTypes[tnponr.Code()] = tnponr
	tnpbns := TrainNotPresentBeforeNextSignal{}
	signalConditionTypes[tnpbns.Code()] = tnpbns
	tnponi := TrainNotPresentOnItems{}
	signalConditionTypes[tnponi.Code()] = tnponi
	tpoi := TrainPresentOnItems{}
	signalConditionTypes[tpoi.Code()] = tpoi
	rs := RouteSet{}
	signalConditionTypes[rs.Code()] = rs
	nsa := NextSignalAspects{}
	signalConditionTypes[nsa.Code()] = nsa
	resa := RouteExitSignalAspects{}
	signalConditionTypes[resa.Code()] = resa
}
