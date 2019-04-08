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
	"time"
)

// A SignalItemManager is in charge of computing the aspect of a signal
type SignalItemManager interface {
	// Name returns a description of this signalItemManager that is used for the UI.
	Name() string
	// GetAspect returns the aspect of the given signal that should be active
	GetAspect(*SignalItem) *SignalAspect
}

// signalLineStyle holds the possible representation shapes for the line at the
// base of the signal.
type signalLineStyle uint8

const (
	lineStyle   signalLineStyle = 0
	bufferStyle signalLineStyle = 1
)

// signalShape holds the possible representation shapes for signal lights.
type signalShape uint8

const (
	noneShape   signalShape = 0
	circleShape signalShape = 1
	// squareShape signalShape = 2
	// QUARTER_SW  signalShape = 10
	// QUARTER_NW  signalShape = 11
	// QUARTER_NE  signalShape = 12
	// QUARTER_SE  signalShape = 13
	// BAR_N_S     signalShape = 20
	// BAR_E_W     signalShape = 21
	// BAR_SW_NE   signalShape = 22
	// BAR_NW_SE   signalShape = 23
	// POLE_NS     signalShape = 31
	// POLE_NSW    signalShape = 32
	// POLE_SW     signalShape = 33
	// POLE_NE     signalShape = 34
	// POLE_NSE    signalShape = 35
)

// ActionTarget defines when a speed limit associated with a signal aspect must be
// applied.
type ActionTarget uint8

// Possible action targets for trains to apply signal actions
const (
	ASAP             ActionTarget = 0
	BeforeThisSignal ActionTarget = 1
	BeforeNextSignal ActionTarget = 2
)

// SignalAction defines an action that must be performed by a train when seeing a
// SignalAspect.
type SignalAction struct {
	Target   ActionTarget
	Speed    float64
	Duration time.Duration
}

// UnmarshalJSON for the SignalAction Type
func (sa *SignalAction) UnmarshalJSON(data []byte) error {
	var rawAction [3]float64
	if err := json.Unmarshal(data, &rawAction); err != nil {
		return fmt.Errorf("unable to read signal action: %s (%s)", data, err)
	}
	*sa = SignalAction{
		Target:   ActionTarget(rawAction[0]),
		Speed:    rawAction[1],
		Duration: time.Duration(rawAction[2]) * time.Second}
	return nil
}

// MarshalJSON for the SignalAction Type
func (sa *SignalAction) MarshalJSON() ([]byte, error) {
	return json.Marshal([3]float64{float64(sa.Target), sa.Speed, float64(sa.Duration)})
}

// SignalAspect class represents an aspect of a signal, that is a combination of
// on and off lights with a meaning for the train driver.
type SignalAspect struct {
	Name         string
	LineStyle    signalLineStyle `json:"lineStyle"`
	OuterShapes  [6]signalShape  `json:"outerShapes"`
	OuterColors  [6]Color        `json:"outerColors"`
	Shapes       [6]signalShape  `json:"shapes"`
	ShapesColors [6]Color        `json:"shapesColors"`
	Actions      []SignalAction  `json:"actions"`
}

// Equals returns true if this aspect is the same as the other aspect,
// i.e. they have the same name.
func (sa *SignalAspect) Equals(other *SignalAspect) bool {
	return sa.Name == other.Name
}

// MeansProceed returns true if this aspect is a proceed aspect, returns false if
// this aspect requires to stop
func (sa *SignalAspect) MeansProceed() bool {
	if len(sa.Actions) == 0 {
		// No actions means the driver discards the signal
		return true
	}
	if sa.Actions[0].Speed != 0 {
		return true
	}
	if sa.Actions[0].Target == BeforeNextSignal {
		return true
	}
	return false
}

// A ConditionType is a type of condition that can be used for defining a signal state.
type ConditionType interface {
	// Code of the ConditionType, uniquely defines this ConditionType
	Code() string
	// SetupTriggers installs needed triggers for the given SignalItem, with the
	// given parameters.
	SetupTriggers(*SignalItem, []string)
	// Solve returns if the condition is met for the given SignalItem and parameters
	Solve(*SignalItem, []string, []string) bool
}

// A Condition on the current simulation context used for defining signal state.
type Condition struct {
	Type   ConditionType
	Values []string
}

// IsMet returns true if this condition is met for the given SignalItem
func (c Condition) IsMet(item *SignalItem, params []string) bool {
	return c.Type.Solve(item, c.Values, params)
}

// A SignalState is an aspect of a signal with a set of conditions to display this
// aspect.
type SignalState struct {
	AspectName string
	Aspect     *SignalAspect
	Conditions map[string]Condition
}

// conditionsMet returns true if all conditions of this SignalState are met (or if
// there is no conditions) on the given signalItem instance.
func (s *SignalState) conditionsMet(signal *SignalItem) bool {
	for _, c := range s.Conditions {
		var params []string
		props, ok := signal.CustomProperties[c.Type.Code()]
		if ok {
			params = props[s.Aspect.Name]
		}
		if !c.IsMet(signal, params) {
			return false
		}
	}
	return true
}

// UnmarshalJSON for the SignalState Type
func (s *SignalState) UnmarshalJSON(data []byte) error {
	var rawSignalState struct {
		AspectName string
		Conditions map[string][]string
	}
	if err := json.Unmarshal(data, &rawSignalState); err != nil {
		return fmt.Errorf("unable to read signal state: %s (%s)", data, err)
	}
	s.AspectName = rawSignalState.AspectName
	if s.Conditions == nil {
		s.Conditions = make(map[string]Condition)
	}
	for k, v := range rawSignalState.Conditions {
		ct, ok := signalConditionTypes[k]
		if !ok {
			return fmt.Errorf("unknown condition type: %s", k)
		}
		s.Conditions[k] = Condition{
			Type:   ct,
			Values: v,
		}
	}
	return nil
}

// MarshalJSON for the SignalState type
func (s *SignalState) MarshalJSON() ([]byte, error) {
	var rawSignalState struct {
		AspectName string
		Conditions map[string][]string
	}
	rawSignalState.AspectName = s.Aspect.Name
	rawSignalState.Conditions = make(map[string][]string)
	for k, v := range s.Conditions {
		rawSignalState.Conditions[k] = v.Values
	}
	return json.Marshal(rawSignalState)
}

// A SignalType describes a type of signals which can have different aspects and
// the logic for displaying aspects.
type SignalType struct {
	Name   string
	States []SignalState
}

// getCustomParams
func (st *SignalType) getDefaultAspect() *SignalAspect {
	if len(st.States) == 0 {
		panic(fmt.Errorf("SignalType %s has no states", st.Name))
	}
	return st.States[len(st.States)-1].Aspect
}

// GetAspect returns the aspect that signal should show according to this SignalType logic.
func (st *SignalType) GetAspect(signal *SignalItem) *SignalAspect {
	for _, state := range st.States {
		if state.conditionsMet(signal) {
			return state.Aspect
		}
	}
	return st.getDefaultAspect()
}

// SignalItem is the "logical" item for signals.
// It holds the logic of a signal defined by its SignalType.
// A signal is the item from and to which routes are created.
type SignalItem struct {
	trackStruct
	Xb             float64 `json:"xn"`
	Yb             float64 `json:"yn"`
	SignalTypeCode string  `json:"signalType"`
	Reverse        bool    `json:"reverse"`
	TrainID        string  `json:"trainID"`

	previousActiveRoute *Route
	nextActiveRoute     *Route
	activeAspect        *SignalAspect
}

// initialize this signalItem
func (si *SignalItem) initialize() error {
	si.activeAspect = si.SignalType().getDefaultAspect()
	return nil
}

// Type returns the name of the type of this item
func (si *SignalItem) Type() TrackItemType {
	return TypeSignal
}

// SignalType returns a pointer to the SignalType of this signal
func (si *SignalItem) SignalType() *SignalType {
	return si.simulation.SignalLib.Types[si.SignalTypeCode]
}

// Reversed() return true if the SignalItem is for trains coming from the right
func (si *SignalItem) Reversed() bool {
	return si.Reverse
}

// BerthOrigin is the Point at which the berth of this signal must be
// displayed by clients. Berths are where train descriptors are displayed.
func (si *SignalItem) BerthOrigin() Point {
	return Point{si.Xb, si.Yb}
}

// ActiveAspect returns the current aspect of the signal
func (si *SignalItem) ActiveAspect() *SignalAspect {
	return si.activeAspect
}

// setActiveRoute sets the given route as active on this SignalItem.
// previous gives the direction.
func (si *SignalItem) setActiveRoute(r *Route, previous TrackItem) {
	si.trackStruct.setActiveRoute(r, previous)
	si.updateSignalState()
}

// IsOnPosition returns true if this signal item is the track item of
// the given position and the position is in the direction of the signal.
func (si *SignalItem) IsOnPosition(pos Position) bool {
	return pos.TrackItem().Equals(si) && pos.PreviousItem().Equals(si.PreviousItem())
}

// Position returns the position of the origin of this signal
func (si *SignalItem) Position() Position {
	return Position{
		simulation:     si.simulation,
		TrackItemID:    si.ID(),
		PreviousItemID: si.PreviousTiID,
		PositionOnTI:   0,
	}
}

// getNextSignal is a helper function that returns the next signal after this one.
//
// If a route starts from this signal, the next signal is the end signal
// of this route. Otherwise, it is the next signal found on the line.
func (si *SignalItem) getNextSignal() *SignalItem {
	if si.nextActiveRoute != nil {
		return si.nextActiveRoute.EndSignal()
	}
	for pos := si.Position(); !pos.IsOut(); pos = pos.Next(DirectionCurrent) {
		if pos.TrackItem().Type() == TypeSignal && pos.TrackItem().IsOnPosition(pos) {
			return pos.TrackItem().(*SignalItem)
		}
	}
	return nil
}

// trainHeadActions performs the actions to be done when a train head reaches this signal item.
//
// In particular, pushes the train code to the next signal.
func (si *SignalItem) trainHeadActions(train *Train) {
	// Check that signal is in same direction as trainHead to push the train
	// descriptor only in this case. For this, we move backwards from the train
	// head to this signal.
	// We do not use isOut, because we are backwards
	for pos := train.TrainHead; pos.TrackItem().Type() != TypeEnd; pos = pos.Previous() {
		if !pos.TrackItem().Equals(si) {
			continue
		}
		if !si.IsOnPosition(pos) {
			// Our signal is the wrong way, so we don't do anything
			return
		}
		if nextSignal := si.getNextSignal(); nextSignal != nil {
			nextSignal.TrainID = train.trainID
		}
		if si.TrainID == train.trainID {
			// Only reset train descriptor if it is ours, as it may
			// be the one of a train behind in the same block
			si.TrainID = ""
		}
	}
	si.updateSignalState()
	si.trackStruct.trainHeadActions(train)
}

// trainTailActions performs the actions to be done when a train tail reaches this signal item.
//
// In particular, deactivate route if auto-cancellable.
func (si *SignalItem) trainTailActions(train *Train) {
	if si.activeRoute != nil &&
		!si.ActiveRoutePreviousItem().Equals(si.PreviousItem()) &&
		!si.activeRoute.BeginSignal().Equals(si) &&
		!si.activeRoute.EndSignal().Equals(si) {
		// The line is highlighted by an opposite direction route or this
		// signal is not the starting/ending signal of this route.
		// => nothing particular to do for this signal
		si.trackStruct.trainTailActions(train)
		return
	}
	// For cleaning purposes: activeRoute not used in this direction
	si.resetActiveRoute()

	if si.previousActiveRoute != nil && si.previousActiveRoute.State != Persistent {
		beginSignalNextRoute := si.previousActiveRoute.BeginSignal().nextActiveRoute
		if beginSignalNextRoute == nil || beginSignalNextRoute.Equals(si.previousActiveRoute) {
			// Only reset previous route if the user did not reactivate it in the meantime
			si.PreviousItem().resetActiveRoute()
			si.resetPreviousActiveRoute(nil)
		}
	}
	if si.nextActiveRoute != nil && si.nextActiveRoute.State != Persistent {
		si.resetNextActiveRoute(nil)
	}
	si.updateSignalState()
	// TODO: trigger previous signal recalculation ?
}

// updateSignalState updates the current signal aspect.
func (si *SignalItem) updateSignalState() {
	oldAspect := si.activeAspect
	switch signalItemManager {
	case nil:
		si.activeAspect = si.SignalType().GetAspect(si)
	default:
		si.activeAspect = signalItemManager.GetAspect(si)
	}
	if !oldAspect.Equals(si.activeAspect) {
		si.simulation.sendEvent(&Event{
			Name:   SignalaspectChanged,
			Object: si,
		})
	}
	if si.previousActiveRoute != nil {
		si.previousActiveRoute.BeginSignal().updateSignalState()
	}
	si.simulation.sendEvent(&Event{
		Name:   TrackItemChanged,
		Object: si,
	})
}

// resetNextActiveRoute information. If route is not nil, do
// this only if the nextActiveRoute is equal to route.
func (si *SignalItem) resetNextActiveRoute(r *Route) {
	if r != nil && si.nextActiveRoute != nil && si.nextActiveRoute.routeID != r.routeID {
		return
	}
	si.nextActiveRoute = nil
	si.updateSignalState()
}

// resetPreviousActiveRoute information. If route is not nil, do
// this only if the previousActiveRoute is equal to route.
func (si *SignalItem) resetPreviousActiveRoute(r *Route) {
	if r != nil && si.previousActiveRoute != nil && si.previousActiveRoute.routeID != r.routeID {
		return
	}
	si.previousActiveRoute = nil
	si.updateSignalState()
}

// MarshalJSON method for SignalItem
func (si *SignalItem) MarshalJSON() ([]byte, error) {
	type jsonSignalItem struct {
		jsonTrackStruct
		Xb                  float64 `json:"xn"`
		Yb                  float64 `json:"yn"`
		SignalTypeCode      string  `json:"signalType"`
		Reverse             bool    `json:"reverse"`
		TrainID             string  `json:"trainID"`
		PreviousActiveRoute string  `json:"previousActiveRoute"`
		NextActiveRoute     string  `json:"nextActiveRoute"`
		ActiveAspect        string  `json:"activeAspect"`
	}
	var parID, narID string
	if si.previousActiveRoute != nil {
		parID = si.previousActiveRoute.ID()
	}
	if si.nextActiveRoute != nil {
		narID = si.nextActiveRoute.ID()
	}
	aSI := jsonSignalItem{
		jsonTrackStruct:     si.asJSONStruct(),
		Xb:                  si.Xb,
		Yb:                  si.Yb,
		SignalTypeCode:      si.SignalTypeCode,
		Reverse:             si.Reverse,
		TrainID:             si.TrainID,
		PreviousActiveRoute: parID,
		NextActiveRoute:     narID,
		ActiveAspect:        si.activeAspect.Name,
	}
	d, err := json.Marshal(aSI)
	return d, err
}

var _ TrackItem = new(SignalItem)

// SignalLibrary holds the information about the signal types and signal aspects
// available in the simulation.
type SignalLibrary struct {
	Aspects map[string]*SignalAspect `json:"signalAspects"`
	Types   map[string]*SignalType   `json:"signalTypes"`
}

// initialize this SignalLibrary
func (sl *SignalLibrary) initialize() error {
	for tName, t := range sl.Types {
		t.Name = tName
		for i, s := range t.States {
			asp, ok := sl.Aspects[s.AspectName]
			if !ok {
				return fmt.Errorf("not aspect with code %s found", s.AspectName)
			}
			asp.Name = s.AspectName
			t.States[i].Aspect = asp
		}
	}
	return nil
}
