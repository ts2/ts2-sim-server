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

package points

import (
	"github.com/ts2/ts2-sim-server/simulation"
	"sync"
)

// StandardManager is a points manager that performs points change
// immediately and never fails.
type StandardManager struct {
	sync.RWMutex
	directions map[string]simulation.PointDirection
}

// Direction returns the direction of the points
func (sm StandardManager) Direction(p *simulation.PointsItem) simulation.PointDirection {
	return sm.directions[p.ID()]
}

// SetDirection tries to set the given PointsItem to the given direction
//
// You should not assume that the direction has been set, since this can be
// delayed or failed. Call Direction to check.
func (sm *StandardManager) SetDirection(p *simulation.PointsItem, dir simulation.PointDirection) {
	if dir == simulation.DirectionCurrent {
		return
	}
	sm.Lock()
	defer sm.Unlock()
	sm.directions[p.ID()] = dir
}

// Name returns a description of this manager that is used for the UI.
func (sm StandardManager) Name() string {
	return "Standard Manager"
}

var _ simulation.PointsItemManager = new(StandardManager)

// newStandardManager returns a pointer to a new StandardManager.
func newStandardManager() *StandardManager {
	return &StandardManager{
		directions: make(map[string]simulation.PointDirection),
	}
}

func init() {
	simulation.RegisterPointsItemManager(newStandardManager())
}
