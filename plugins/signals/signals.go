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

package signals

import "github.com/ts2/ts2-sim-server/simulation"

// StandardManager is a lines manager that never fails.
type StandardManager struct{}

// Name returns a description of this signalItemManager that is used for the UI.
func (sm StandardManager) Name() string {
	return "Standard Manager"
}

// GetAspect returns the aspect of the given signal that should be active
func (sm StandardManager) GetAspect(signal *simulation.SignalItem) *simulation.SignalAspect {
	// Don't do anything particular in this implementation (no failures)
	return signal.SignalType().GetAspect(signal)
}

var _ simulation.SignalItemManager = StandardManager{}

func init() {
	simulation.RegisterSignalItemManager(StandardManager{})
}
