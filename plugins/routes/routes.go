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

package routes

import (
	"fmt"

	"github.com/ts2/ts2-sim-server/simulation"
)

// StandardManager is a routes manager that allow activation
// by checking if routes are conflicting or not.
// It always allows deactivation.
type StandardManager struct{}

// CanActivate returns an error if the given route cannot be activated.
// In this implementation, it checks route conflicts and returns
// false if a conflict is found.
func (sm StandardManager) CanActivate(r *simulation.Route) error {
	var flag *simulation.Route
	for _, pos := range r.Positions {
		if pos.TrackItem().ID() == r.BeginSignalId || pos.TrackItem().ID() == r.EndSignalId {
			continue
		}
		if pos.TrackItem().ConflictItem() != nil && pos.TrackItem().ConflictItem().ActiveRoute() != nil {
			// Our trackItem has a conflicting item with an active route
			return fmt.Errorf("conflicting route %s is active", pos.TrackItem().ConflictItem().ActiveRoute().ID())
		}
		if pos.TrackItem().ActiveRoute() == nil {
			if flag != nil {
				// We had a route with same direction but does not end with the same signal
				return fmt.Errorf("conflicting route %s is active", flag.ID())
			}
			continue
		}
		// The track item has an active route already
		if pos.TrackItem().Type() == simulation.TypePoints && flag == nil {
			// The trackItem is a pointsItem and it is the first
			// trackItem with active route that we meet
			return fmt.Errorf("conflicting route %s is active", pos.TrackItem().ActiveRoute().ID())
		}
		if pos.PreviousItem().ID() != pos.TrackItem().ActiveRoutePreviousItem().ID() {
			// The direction of route r is different from that of the active route of the TI
			return fmt.Errorf("conflicting route %s is active", pos.TrackItem().ActiveRoute().ID())
		}
		if pos.TrackItem().ActiveRoute().ID() == r.ID() {
			// Always allow to setup the same route again
			return nil
		}
		// We set flag to true to remember we have come across an item with activeRoute with
		// the same direction. This enables the user to set a route ending with the same end
		// signal when it is cleared by a train still on the route
		flag = pos.TrackItem().ActiveRoute()
	}
	return nil
}

// CanDeactivate returns an error if the given route cannot be deactivated.
// In this implementation, it always returns true.
func (sm StandardManager) CanDeactivate(r *simulation.Route) error {
	return nil
}

// Name of this manager
func (sm StandardManager) Name() string {
	return "Standard Manager"
}

var _ simulation.RoutesManager = StandardManager{}

func init() {
	simulation.RegisterRoutesManager(StandardManager{})
}
