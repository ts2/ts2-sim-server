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

package trains

import (
	"fmt"
	"math"
	"time"

	"github.com/ts2/ts2-sim-server/simulation"
)

const (
	defaultMaxDistance float64 = 50
	lineSafetyDistance float64 = 100
)

// The StandardManager implements a driver behaviour where trains
// accelerate as soon as possible to the maximum possible speed and
// brake at the latest to keep in speed limits.
type StandardManager struct{}

// Speed computes and returns the speed of the given train after timeElapsed
func (m StandardManager) Speed(t *simulation.Train, timeElapsed time.Duration) float64 {
	if !t.IsActive() || t.Status == simulation.Stopped {
		return 0
	}
	// secs is the time elapsed in seconds
	secs := float64(timeElapsed) / float64(time.Second)

	// maxDistance is the maximum distance we look ahead to find speed limits
	maxDistance := math.Max(math.Pow(t.Speed, 2)/t.TrainType().StdBraking, defaultMaxDistance)

	// Get distances to next targets
	dtnStation, okStation := distanceToNextStop(t, maxDistance)
	dtnSpeedLimit, speedLimit, okSpeedLimit := nextSpeedLimit(t, maxDistance, secs)
	dtnTrain, okTrain := distanceToNextTrain(t, maxDistance)
	safetyDistance := lineSafetyDistance
	if t.IsShunting() {
		safetyDistance = 0
	}
	if okTrain {
		dtnTrain -= safetyDistance
	}
	// Get distance to next signal depending on actions
	nsp := t.NextSignalPosition()
	dtnSignal, okSignal := distanceToNextSignal(t)
	switch t.ApplicableAction().Target {
	case simulation.ASAP:
		// We emulate a distance to next signal to get a stdBraking
		dtnSignal = (math.Pow(t.Speed-t.TrainType().StdBraking*secs, 2)-math.Pow(t.ApplicableAction().Speed, 2))/
			(2*t.TrainType().StdBraking) + (t.Speed * secs / 2)
	case simulation.BeforeNextSignal:
		if nsp.TrackItemID == t.LastSeenSignal().ID() {
			// The signal with the applicable action is still ahead
			nnSignalPos := simulation.NextSignalPosition(nsp)
			if !nnSignalPos.IsNull() {
				extraDistance, _ := nnSignalPos.Sub(nsp)
				dtnSignal += extraDistance
			} else {
				okSignal = false
				dtnSignal = 0
			}
		}
	}

	// Calculate speeds to manage each target
	targetSpeedForStation := getMaxSpeed(t)
	if okStation {
		targetSpeedForStation = targetSpeed(t, secs, dtnStation, 0)
	}
	targetSpeedForLimit := getMaxSpeed(t)
	if okSpeedLimit {
		targetSpeedForLimit = targetSpeed(t, secs, dtnSpeedLimit, speedLimit)
	}
	targetSpeedForTrain := getMaxSpeed(t)
	if okTrain {
		targetSpeedForTrain = targetSpeed(t, secs, dtnTrain, 0)
	}
	targetSpeedForSignal := getMaxSpeed(t)
	if okSignal {
		targetSpeedForSignal = targetSpeed(t, secs, dtnSignal, t.ApplicableAction().Speed)
	}
	if t.ApplicableAction().Target == simulation.BeforeThisSignal && nsp.TrackItemID != t.LastSeenSignal().ID() {
		// We passed the signal, and we keep its speed limit until we see the next one.
		targetSpeedForSignal = t.ApplicableAction().Speed
	}
	targetSpeed := math.Min(targetSpeedForStation,
		math.Min(targetSpeedForLimit,
			math.Min(targetSpeedForTrain, targetSpeedForSignal)))
	acceleration := math.Max(-t.TrainType().EmergBraking,
		math.Min(1/secs*(targetSpeed-t.Speed), t.TrainType().StdAccel))
	simulation.Logger.Debug("Set Train speed", "ID", t.ID(),
		"dtnStation", dtnStation,
		"dtnSpeedLimit", dtnSpeedLimit,
		"dtnTrain", dtnTrain,
		"dtnSignal", dtnSignal,
		"targetSpeedForStation", targetSpeedForStation,
		"targetSpeedForLimit", targetSpeedForLimit,
		"targetSpeedForTrain", targetSpeedForTrain,
		"targetSpeedForSignal", targetSpeedForSignal)
	return math.Max(0, t.Speed+acceleration*secs)
}

// Name of this manager, for use in UI messages
func (m StandardManager) Name() string {
	return "Standard Manager"
}

// distanceToNextSignal returns the distance to the next signal by looking forward of
// the given train's head
//
// Second parameter is true if a signal has been found.
func distanceToNextSignal(t *simulation.Train) (float64, bool) {
	nsp := t.NextSignalPosition()
	if nsp.IsNull() {
		return 0, false
	}
	dtns, err := nsp.Sub(t.TrainHead)
	if err != nil {
		panic(fmt.Sprintf("unexpected error: %s", err))
	}
	return dtns, true
}

// distanceToNextStop returns the distance to the next stop by looking forward of
// the given train's head up to a maximum distance of maxDistance.
//
// Second parameter is true if a stop has been found.
func distanceToNextStop(t *simulation.Train, maxDistance float64) (float64, bool) {
	if t.Service() == nil || t.NextPlaceIndex == simulation.NoMorePlace {
		// No service assigned or no place to call at
		return 0, false
	}
	var line *simulation.ServiceLine
	for i := t.NextPlaceIndex; i < len(t.Service().Lines); i++ {
		if t.Service().Lines[i].MustStop {
			line = t.Service().Lines[i]
			break
		}
	}
	if line == nil {
		// No more stops for this service
		return 0, false
	}
	pos := t.TrainHead
	distance := pos.TrackItem().RealLength() - t.TrainHead.PositionOnTI
	for pos.TrackItem().Type() != simulation.TypeEnd && distance < maxDistance {
		ti := pos.TrackItem()
		if ti.Type() == simulation.TypeSignal && ti.IsOnPosition(pos) && !ti.(*simulation.SignalItem).ActiveAspect().MeansProceed() {
			// We have found a red signal, no need to go further
			return 0, false
		}
		if ti.Place() == line.Place() {
			return distance, true
		}
		pos = pos.Next(simulation.DirectionCurrent)
		distance += pos.TrackItem().RealLength()
	}
	return 0, false
}

// nextSpeedLimit returns the distance and the value of the next speed limit by looking forward of
// the given train's head up to a maximum distance of maxDistance.
//
// The last argument is true if a new speed limit has been found within maxDistance.
func nextSpeedLimit(t *simulation.Train, maxDistance, secs float64) (float64, float64, bool) {
	pos := t.TrainHead
	distance := pos.TrackItem().RealLength() - t.TrainHead.PositionOnTI
	for pos.TrackItem().Type() != simulation.TypeEnd && distance < maxDistance {
		pos = pos.Next(simulation.DirectionCurrent)
		ti := pos.TrackItem()
		if ti.MaxSpeed() < getMaxSpeed(t)-t.TrainType().StdBraking*secs {
			return distance, ti.MaxSpeed(), true
		}
		distance += ti.RealLength()
	}
	return 0, getMaxSpeed(t), false
}

// distanceToNextTrain returns the distance to the next train by looking forward of
// the given train's head up to a maximum distance of maxDistance.
//
// Second parameter is true if a train has been found.
func distanceToNextTrain(t *simulation.Train, maxDistance float64) (float64, bool) {
	pos := t.TrainHead
	var distance float64
	for pos.TrackItem().Type() != simulation.TypeEnd && distance < maxDistance {
		ti := pos.TrackItem()
		if ti.Type() == simulation.TypeSignal && ti.IsOnPosition(pos) && !ti.(*simulation.SignalItem).ActiveAspect().MeansProceed() {
			// We have found a red signal, no need to go further
			return 0, false
		}
		distanceToTrain, ok := ti.DistanceToTrainEnd(pos)
		if ok {
			return distance + distanceToTrain, true
		}
		pos = pos.Next(simulation.DirectionCurrent)
		if distance == 0 {
			distance = ti.RealLength() - t.TrainHead.PositionOnTI
			continue
		}
		distance += ti.RealLength()
	}
	return 0, false
}

// targetSpeed defines the current target speed of the train depending on the parameters.
func targetSpeed(t *simulation.Train, secs, targetDistance, targetSpeed float64) float64 {
	// d is the maximum distance that can be travelled during the last
	// sample. It is used to determine when to stop the train.
	d := 0.5 * t.TrainType().StdBraking * math.Pow(secs, 2)
	if targetDistance < d {
		return targetSpeed
	}

	theoreticalSpeed := calculatedSpeed(t, targetDistance, targetSpeed)

	// s1 is half the distance run at the train's current speed during secs
	// This value is used to get a centered sampling of the braking curve.
	s1 := t.Speed * secs / 2
	// s2 is equivalent to s1, but taking into account the theoreticalSpeed
	s2 := theoreticalSpeed * secs / 2

	if theoreticalSpeed < t.Speed {
		return calculatedSpeed(t, targetDistance-s1, targetSpeed)
	}
	return calculatedSpeed(t, targetDistance-s2, targetSpeed)
}

// calculatedSpeed returns the speed the train should be right now to be able to be
// at a speed of targetSpeedAtPos at a distance of targetDistance from
// the train head, not exceeding maxSpeed. This function does not take
// into account any sampling margin.
func calculatedSpeed(t *simulation.Train, targetDistance, targetSpeed float64) float64 {
	return math.Min(
		getMaxSpeed(t),
		math.Sqrt(math.Abs(2*targetDistance*t.TrainType().StdBraking)+math.Pow(targetSpeed, 2)))
}

// getMaxSpeed returns the maximum speed allowed for the train in its current position
func getMaxSpeed(t *simulation.Train) float64 {
	return math.Min(t.TrainType().MaxSpeed, t.TrainHead.TrackItem().MaxSpeed())
}

var _ simulation.TrainsManager = StandardManager{}

func init() {
	simulation.RegisterTrainsManager(StandardManager{})
}
