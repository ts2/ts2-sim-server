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
	"math"
	"time"
)

// A TrainsManager defines a driver behaviour which impacts the speed of trains.
type TrainsManager interface {
	// Speed computes and returns the speed of the given train after timeElapsed
	Speed(*Train, time.Duration) float64
	// Name of this manager used for UI messages
	Name() string
}

// The TrainStatus describe the current state of a train
type TrainStatus uint8

const (
	// Inactive means not yet entered on the scene
	Inactive TrainStatus = 0

	// Running with a positive speed
	Running TrainStatus = 10

	// Stopped at a station
	Stopped TrainStatus = 20

	// Waiting means an unscheduled stop, e.g. at a red signal
	Waiting TrainStatus = 30

	// Out means the train exited the area
	Out TrainStatus = 40

	// EndOfService means the train has finished its service and no new service assigned
	EndOfService TrainStatus = 50
)

// VeryHighSpeed is the speed limit set when there are no speed limits.
// It is higher than the highest possible train speed ever.
const VeryHighSpeed = 999

// minRunningSpeed is the minimum speed at which a train is considered running
const minRunningSpeed float64 = 0.25

// Train is a stock of `TrainType` running on a track at a certain speed and to which
// is assigned a `Service`.
type Train struct {
	trainID        string         `json:"-"`
	AppearTime     Time           `json:"appearTime"`
	InitialDelay   DelayGenerator `json:"initialDelay"`
	InitialSpeed   float64        `json:"initialSpeed"`
	NextPlaceIndex int            `json:"nextPlaceIndex"`
	ServiceCode    string         `json:"serviceCode"`
	Speed          float64        `json:"speed"`
	Status         TrainStatus    `json:"status"`
	StoppedTime    int            `json:"stoppedTime"`
	TrainTypeCode  string         `json:"trainTypeCode"`
	TrainHead      Position       `json:"trainHead"`

	trainManager    TrainsManager
	simulation      *Simulation
	effInitialDelay time.Duration
	minStopTime     time.Duration
	signalActions   []SignalAction
	actionIndex     int
	actionTime      Time
	lastSignal      *SignalItem
}

// ID returns the unique internal identifier of this Train
func (t *Train) ID() string {
	return t.trainID
}

// setSimulation sets the simulation to this Train
func (t *Train) setSimulation(sim *Simulation) {
	t.simulation = sim
	t.TrainHead.simulation = sim
}

// initialize attaches the Simulation to this Train and initializes it.
func (t *Train) initialize(id string) {
	t.trainID = id
	t.effInitialDelay = t.InitialDelay.Yield()
	if t.InitialDelay.IsNull() {
		t.effInitialDelay = t.simulation.Options.DefaultDelayAtEntry.Yield()
	}
	t.minStopTime = t.simulation.Options.DefaultMinimumStopTime.Yield()
	if t.trainManager == nil {
		t.trainManager = defaultTrainManager
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

// TrainTail returns the position of the train's tail
func (t *Train) TrainTail() Position {
	return t.TrainHead.Add(-t.TrainType().Length)
}

// MarshalJSON method for the train type
func (t Train) MarshalJSON() ([]byte, error) {
	type auxTrain Train
	type trainJSON struct {
		auxTrain
		ID string `json:"id"`
	}
	at := trainJSON{
		auxTrain: auxTrain(t),
		ID:       t.ID(),
	}
	return json.Marshal(at)
}

// IsActive returns true if this train is in the area and its service is not finished.
func (t *Train) IsActive() bool {
	return t.Status != Inactive &&
		t.Status != Out &&
		t.Status != EndOfService
}

// activate this Train if this train is Inactive and if h is after its AppearTime.
//
// In all other cases, this method is a no-op
func (t *Train) activate(h Time) {
	if t.Status != Inactive {
		return
	}
	if h.IsZero() {
		return
	}
	realAppearTime := t.AppearTime.Add(t.effInitialDelay)
	if h.Sub(realAppearTime) < 0 {
		return
	}
	t.Speed = t.InitialSpeed
	// Update signals
	if signalAhead := t.findNextSignal(); signalAhead != nil {
		signalAhead.setTrain(t)
	}
	// Status update
	t.Status = Running
	if t.StoppedTime != 0 || t.Service() == nil {
		t.Status = Stopped
	}
	if t.Service() != nil {
		t.NextPlaceIndex = 0
	}
	t.executeActions(0)
	// Signal actions update
	t.signalActions = []SignalAction{{
		Target: ASAP,
		Speed:  VeryHighSpeed,
	}}
	t.actionIndex = 0
	// Log status change
	t.logTrainEntersArea()
}

// advance the train by a step corresponding to the elapsed time,
// and executes all the associated actions.
func (t *Train) advance(timeElapsed time.Duration) {
	if !t.IsActive() {
		return
	}
	t.updateSignalActions()
	t.Speed = t.trainManager.Speed(t, timeElapsed)
	advanceLength := t.Speed * float64(timeElapsed) / float64(time.Second)
	t.TrainHead = t.TrainHead.Add(advanceLength)
	t.updateStatus(timeElapsed)
	t.executeActions(advanceLength)
	t.simulation.sendEvent(&Event{
		Name:   TrainChanged,
		Object: t,
	})
}

// Execute actions that have to be done when the train head enters
// a trackItem or when the train tail leaves another.
//
// For each case this is done in two stages:
//
//   - first execute actions related to the train itself
//   - then call TrackItem.trainHeadActions() or TrackItem.trainTailActions()).
func (t *Train) executeActions(advanceLength float64) {
	// Train head
	oth := t.TrainHead.Add(-advanceLength)
	toNotify := make(map[TrackItem]bool)
	for _, ti := range oth.trackItemsToPosition(t.TrainHead) {
		t.checkPlace(ti)
		t.updateItemWithTrainHead(ti)
		ti.trainHeadActions(t)
		toNotify[ti] = true
	}
	// Train tail
	tt := t.TrainTail()
	ott := tt.Add(-advanceLength)
	for _, ti := range ott.trackItemsToPosition(tt) {
		t.updateItemWithTrainTail(ti)
		ti.trainTailActions(t)
		toNotify[ti] = true
	}
	if tt.IsOut() {
		t.Status = Out
		t.Speed = 0
		t.logAndScoreTrainExited()
	}
	for ti := range toNotify {
		t.simulation.sendEvent(&Event{
			Name:   TrackItemChanged,
			Object: ti,
		})
	}
}

// updateItemWithTrainHead updates the knowledge of this trackItem about this train's Head,
// knowing that this item is between the former head and the current head of the train.
func (t *Train) updateItemWithTrainHead(ti TrackItem) {
	ti.underlying().trainEndsFW[t] = ti.RealLength()
	ti.underlying().trainEndsBK[t] = 0
	if t.simulation.Options.TrackCircuitBased {
		return
	}
	if ti.Equals(t.TrainHead.TrackItem()) {
		if t.TrainHead.PreviousItemID == ti.PreviousItem().ID() {
			ti.underlying().trainEndsFW[t] = t.TrainHead.PositionOnTI
		} else {
			ti.underlying().trainEndsBK[t] = ti.RealLength() - t.TrainHead.PositionOnTI
		}
	}
}

// updateItemWithTrainTail updates the knowledge of this trackItem about this train's Tail,
// knowing that this item is between the former tail and the current tail of the train.
func (t *Train) updateItemWithTrainTail(ti TrackItem) {
	if !ti.Equals(t.TrainHead.TrackItem()) {
		delete(ti.underlying().trainEndsBK, t)
		delete(ti.underlying().trainEndsFW, t)
	}
	if ti.Equals(t.TrainTail().TrackItem()) {
		if t.simulation.Options.TrackCircuitBased {
			ti.underlying().trainEndsFW[t] = ti.RealLength()
			ti.underlying().trainEndsBK[t] = 0
			return
		}
		if t.TrainTail().PreviousItemID == ti.PreviousItem().ID() {
			ti.underlying().trainEndsBK[t] = t.TrainTail().PositionOnTI
		} else {
			ti.underlying().trainEndsFW[t] = ti.RealLength() - t.TrainTail().PositionOnTI
		}
		if !ti.Equals(t.TrainHead.TrackItem()) {
			if t.TrainTail().PreviousItemID == ti.PreviousItem().ID() {
				ti.underlying().trainEndsFW[t] = ti.RealLength()
			} else {
				ti.underlying().trainEndsBK[t] = 0
			}
		}
	}
}

// NextSignalPosition returns the position of the next signal in front of this train
//
// Returns a null position if there is no signal ahead.
func (t *Train) NextSignalPosition() Position {
	return NextSignalPosition(t.TrainHead)
}

// NextSignalPosition returns the position of the next signal in front of the given position.
//
// Returns a null position if there is no signal ahead.
func NextSignalPosition(pos Position) Position {
	if pos.TrackItem().Type() == TypeEnd {
		return Position{}
	}
	cur := pos.Next(DirectionCurrent)
	for cur.TrackItem().Type() != TypeEnd {
		if cur.TrackItem().Type() == TypeSignal && cur.TrackItem().(*SignalItem).IsOnPosition(cur) {
			return cur
		}
		cur = cur.Next(DirectionCurrent)
	}
	return Position{}

}

// findNextSignal returns the next signal in front of this Train
func (t *Train) findNextSignal() *SignalItem {
	return t.NextSignalPosition().TrackItem().(*SignalItem)
}

// updateSignalActions updates the applicable signal actions list based on the position
// of the train and the visible signal.
func (t *Train) updateSignalActions() {
	nsp := t.NextSignalPosition()
	if nsp.Equals(Position{}) {
		// No more signal ahead
		t.signalActions = []SignalAction{{
			Target: ASAP,
			Speed:  VeryHighSpeed,
		}}
		t.actionIndex = 0
		return
	}
	nsd, err := nsp.Sub(t.TrainHead)
	if err != nil {
		Logger.Crit("unexpected error", "error", err)
		return
	}
	if nsd < t.simulation.Options.DefaultSignalVisibility {
		// We can see the next signal aspect
		if len(nsp.TrackItem().(*SignalItem).activeAspect.Actions) > 0 {
			// It requires actions
			// We check actions each time because the aspect of the signal
			// might have changed
			t.signalActions = nsp.TrackItem().(*SignalItem).activeAspect.Actions
			if t.lastSignal == nil || !nsp.TrackItem().Equals(t.lastSignal) {
				// We see this signal for the first time
				t.lastSignal = nsp.TrackItem().(*SignalItem)
				t.actionIndex = 0
				t.actionTime = Time{}
			}
		} else {
			// This signal does not require actions, so we only update our
			// memory of the last signal
			t.lastSignal = nsp.TrackItem().(*SignalItem)
		}
	}

	currentTime := t.simulation.Options.CurrentTime
	if math.Abs(t.Speed-t.ApplicableAction().Speed) < 0.1 {
		// We have achieved the action's target speed.
		if t.actionTime.IsZero() {
			// Start the waiting time
			t.actionTime = currentTime
		}
		if currentTime.After(t.actionTime.Add(t.ApplicableAction().Duration)) {
			// We have waited enough, so we go to next action if any
			if len(t.signalActions) > t.actionIndex+1 {
				t.actionIndex += 1
			}
		}
	}
}

// checkPlace if the given ti belongs to a place which is a waypoiny on t's service (non stop).
// Updates t's current service line accordingly.
func (t *Train) checkPlace(ti TrackItem) {
	if ti.Type() != TypeLine && ti.Type() != TypeInvisibleLink {
		return
	}
	if ti.Place() == nil {
		return
	}
	if t.Service() == nil || t.NextPlaceIndex == NoMorePlace {
		return
	}
	sLine := t.Service().Lines[t.NextPlaceIndex]
	if sLine.PlaceCode != ti.underlying().PlaceCode {
		return
	}
	if sLine.MustStop {
		// If we must stop, then we will change the current line at departure
		return
	}
	t.jumpToNextServiceLine()
}

// jumpToNextServiceLine sets the next service line as the new active line.
func (t *Train) jumpToNextServiceLine() {
	t.minStopTime = t.simulation.Options.DefaultMinimumStopTime.Yield()
	if t.NextPlaceIndex == len(t.Service().Lines)-1 {
		// The service is ended
		t.NextPlaceIndex = NoMorePlace
		for _, action := range t.Service().PostActions {
			switch action.ActionCode {
			case actionReverse:
				t.reverse()
			case actionSetService:
				t.ServiceCode = action.ActionParam
				t.NextPlaceIndex = 0
				t.findNextSignal().setTrain(t)
				if t.StoppedTime != 0 {
					t.Status = Stopped
				} else {
					t.Status = Running
				}
			}
		}
		return
	}
	t.NextPlaceIndex += 1
}

// reverse the train direction
func (t *Train) reverse() {
	if t.Speed != 0 {
		return
	}
	if signalAhead := t.findNextSignal(); signalAhead != nil {
		signalAhead.setTrain(nil)
	}
	if activeRoute := t.TrainHead.TrackItem().ActiveRoute(); activeRoute != nil {
		if err := activeRoute.Deactivate(); err != nil {
			t.simulation.MessageLogger.addMessage(err.Error(), simulationMsg)
		}
	}
	t.TrainHead = t.TrainTail().Reversed()
	if newSignalAhead := t.findNextSignal(); newSignalAhead != nil {
		newSignalAhead.setTrain(t)
	}
	t.signalActions = []SignalAction{{
		Target: ASAP,
		Speed:  VeryHighSpeed,
	}}
	t.actionIndex = 0
	t.updateSignalActions()
}

// IsShunting returns true if this train is currently shunting.
func (t *Train) IsShunting() bool {
	return false
}

// ApplicableAction returns the current signal action that this train is following
func (t *Train) ApplicableAction() SignalAction {
	return t.signalActions[t.actionIndex]
}

// LastSeenSignal returns the last signal seen by the driver. It may still be in front
// of the train head.
func (t *Train) LastSeenSignal() *SignalItem {
	return t.lastSignal
}

// updateStatus of the train
func (t *Train) updateStatus(timeElapsed time.Duration) {
	if !t.IsActive() {
		return
	}
	if t.Speed > minRunningSpeed {
		// Speed is not null, the train is running
		t.Status = Running
		return
	}
	if t.Service() == nil || t.NextPlaceIndex == NoMorePlace {
		// Train is stopped but not assigned any service
		t.Status = Waiting
		return
	}
	// The train is operating on a service that is not over
	line := t.Service().Lines[t.NextPlaceIndex]
	thi := t.TrainHead.TrackItem()
	if thi.Type() != TypeLine || thi.Place() == nil || thi.Place().PlaceCode != line.PlaceCode {
		// Train is stopped but not assigned any service
		t.Status = Waiting
		return
	}
	// Train is stopped at the scheduled nextStop place
	if t.Status == Running {
		// Train just stopped
		t.Status = Stopped
		t.StoppedTime = 0
		t.simulation.sendEvent(&Event{
			Name:   TrainStoppedAtStationEvent,
			Object: t,
		})
		t.logAndScoreTrainStoppedAtStation()
		return
	}
	if t.Status != Stopped {
		// Typically end of service
		return
	}
	// Train is already stopped at the place
	if line.ScheduledDepartureTime.Sub(t.simulation.Options.CurrentTime) > 0 ||
		t.StoppedTime < int(t.minStopTime/time.Second) ||
		line.ScheduledDepartureTime.IsZero() {
		// Conditions to depart are not met
		t.Status = Stopped
		t.StoppedTime += int(timeElapsed / time.Second)
		return
	}
	// Train departs
	oldServiceCode := t.ServiceCode
	t.jumpToNextServiceLine()
	if oldServiceCode != t.ServiceCode {
		// The service has changed
		if t.TrainHead.TrackItem().Place().PlaceCode != t.Service().Lines[t.NextPlaceIndex].PlaceCode {
			// The first scheduled place of this new service is not here, so we depart
			t.Status = Running
			t.simulation.sendEvent(&Event{
				Name:   TrainDepartedFromStationEvent,
				Object: t,
			})
			return
		}
		// This is also the first scheduled place of the new service
		t.Status = Stopped
		return
	}
	if t.NextPlaceIndex == NoMorePlace {
		// No other place to call at
		t.Status = EndOfService
		return
	}
	// There are still places to call at
	t.Status = Running
	t.simulation.sendEvent(&Event{
		Name:   TrainDepartedFromStationEvent,
		Object: t,
	})
}

// logTrainEntersArea sends a message on the logger saying that this train entered
// the area and informing if it is late or early.
func (t *Train) logTrainEntersArea() {
	var msg string
	switch {
	case -time.Minute < t.effInitialDelay && t.effInitialDelay < time.Minute:
		msg = fmt.Sprintf("Train %s entered the area on time", t.ServiceCode)
	case t.effInitialDelay <= -60:
		msg = fmt.Sprintf("Train %s entered the area %d minutes early", t.ServiceCode, t.effInitialDelay/time.Minute)
	case t.effInitialDelay >= 60:
		msg = fmt.Sprintf("Train %s entered the area %d minutes late", t.ServiceCode, t.effInitialDelay/time.Minute)
	}
	t.simulation.MessageLogger.addMessage(msg, simulationMsg)
}

// logAndScoreTrainStoppedAtStation modifies the score and logs information about this train
// that has just stopped at a station.
func (t *Train) logAndScoreTrainStoppedAtStation() {
	serviceLine := t.Service().Lines[t.NextPlaceIndex]
	place := t.TrainHead.TrackItem().Place()
	plannedPlatform := serviceLine.TrackCode
	actualPlatform := t.TrainHead.TrackItem().TrackCode()
	sim := t.simulation
	if actualPlatform != plannedPlatform {
		sim.updateScore(sim.Options.WrongPlatformPenalty)
		sim.MessageLogger.addMessage(fmt.Sprintf("Train %s arrived at station %s on platform %s instead of %s",
			t.ServiceCode, place.Name(), actualPlatform, plannedPlatform), simulationMsg)
	}
	scheduledArrivalTime := serviceLine.ScheduledArrivalTime
	currentTime := sim.Options.CurrentTime
	delay := currentTime.Sub(scheduledArrivalTime)
	if delay > time.Minute {
		playerDelay := delay - t.effInitialDelay
		if playerDelay > time.Minute {
			sim.updateScore(sim.Options.LatePenalty * int(playerDelay/time.Minute))
		}
		sim.MessageLogger.addMessage(fmt.Sprintf("Train %s arrived %d minutes late at station %s (%+d minutes)",
			t.ServiceCode, delay/time.Minute, place.Name(), playerDelay/time.Minute), simulationMsg)
		return
	}
	sim.MessageLogger.addMessage(fmt.Sprintf("Train %s arrived on time at station %s",
		t.ServiceCode, place.Name()), simulationMsg)
}

// logAndScoreTrainExited modifies the score and logs information about this train
// that just exited the area.
func (t *Train) logAndScoreTrainExited() {
	sim := t.simulation
	if t.NextPlaceIndex != NoMorePlace {
		sim.updateScore(sim.Options.WrongDestinationPenalty)
		sim.MessageLogger.addMessage(fmt.Sprintf("Train %s badly routed", t.ServiceCode), simulationMsg)
	}
	sim.MessageLogger.addMessage(fmt.Sprintf("Train %s exited the area", t.ServiceCode), simulationMsg)
}
