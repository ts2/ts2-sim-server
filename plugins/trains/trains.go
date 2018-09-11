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
	"math"
	"time"

	"github.com/ts2/ts2-sim-server/simulation"
)

const defaultMaxDistance float64 = 50

// The StandardManager implements a driver behaviour where trains
// accelerate as soon as possible to the maximum possible speed and
// brake at the latest to keep in speed limits.
type StandardManager struct{}

// Speed computes and returns the speed of the given train after timeElapsed
func (m StandardManager) Speed(t *simulation.Train, timeElapsed time.Duration) float64 {
	if !t.IsActive() || t.Status == simulation.Stopped {
		return 0
	}
	// k is the gain factor to set acceleration from the difference
	// between current speed and target speed
	k := float64(time.Second) / float64(timeElapsed)

	// maxDistance is the maximum distance we look ahead to find speed limits
	maxDistance := math.Max(math.Pow(t.Speed, 2)/t.TrainType().StdBraking, defaultMaxDistance)

	// Get distances to next targets
	dtnStation, okStation := m.getDistanceToNextStop(t, maxDistance)
	dtnSpeedLimit, speedLimit, okSpeedLimit := m.getNextSpeedLimit(t, maxDistance, k)

}

// Name of this manager, for use in UI messages
func (m StandardManager) Name() string {
	return "StandardManager"
}

// getDistanceToNextStop returns the distance to the next stop by looking forward of
// the given trains head up to a maximum distance of maxDistance.
//
// Second parameter is true if a stop has been found.
func (m StandardManager) getDistanceToNextStop(t *simulation.Train, maxDistance float64) (float64, bool) {
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
		if ti.Place().ID() == line.Place().ID() {
			return distance, true
		}
		pos = pos.Next(simulation.DirectionCurrent)
		distance += pos.TrackItem().RealLength()
	}
	return 0, false
}

// getNextSpeedLimit returns the distance and the value of the next speed limit.
// The last argument is true if a new speed limit has been found within maxDistance.
func (m StandardManager) getNextSpeedLimit(t *simulation.Train, maxDistance, k float64) (float64, float64, bool) {
	pos := t.TrainHead
	distance := pos.TrackItem().RealLength() - t.TrainHead.PositionOnTI
	for pos.TrackItem().Type() != simulation.TypeEnd && distance < maxDistance {
		pos = pos.Next(simulation.DirectionCurrent)
		ti := pos.TrackItem()
		if ti.MaxSpeed() < m.getMaxSpeed(t)-t.TrainType().StdBraking/k {
			return distance, ti.MaxSpeed(), true
		}
		distance += ti.RealLength()
	}
	return 0, m.getMaxSpeed(t), false
}

// getMaxSpeed returns the maximum speed allowed for the train in its current position
func (m StandardManager) getMaxSpeed(t *simulation.Train) float64 {
	return math.Min(t.TrainType().MaxSpeed, t.TrainHead.TrackItem().MaxSpeed())
}

var _ simulation.TrainsManager = StandardManager{}

func init() {
	simulation.RegisterTrainsManager(StandardManager{})
}
