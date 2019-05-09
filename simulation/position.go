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
	"errors"
	"fmt"
)

// A Position object is a point on a TrackItem.
//
// A Position is defined as being positionOnTI meters away from the end of this
// TrackItem that is connected to PreviousItem.
//
// Note that a Position has a direction, so that for any point on a TrackItem,
// there are two Positions that can be defined:
//
//  - one starting from one end of the TrackItem.
//  - the other starting from the other end.
//
// You can get the other Position by calling Reversed()
type Position struct {
	simulation *Simulation

	TrackItemID    string  `json:"trackItem"`
	PreviousItemID string  `json:"previousTI"`
	PositionOnTI   float64 `json:"positionOnTI"`
}

// TrackItem of this Position
func (pos Position) TrackItem() TrackItem {
	return pos.simulation.TrackItems[pos.TrackItemID]
}

// PreviousItem of this Position
func (pos Position) PreviousItem() TrackItem {
	return pos.simulation.TrackItems[pos.PreviousItemID]
}

// IsValid returns true if this is a valid position (i.e. items are connected, and
// distance is positive), false otherwise.
func (pos Position) IsValid() bool {
	if pos.IsNull() {
		// A null position with simulation is valid
		return true
	}
	if pos.simulation == nil {
		return false
	}
	if pos.TrackItemID == "" {
		return false
	}
	if pos.PositionOnTI > pos.TrackItem().RealLength() || pos.PositionOnTI < 0 {
		return false
	}
	if !pos.TrackItem().IsConnected(pos.PreviousItem()) {
		return false
	}
	return true
}

// IsNull returns true if this Position is null.
func (pos Position) IsNull() bool {
	return pos.TrackItemID == "" &&
		pos.PreviousItemID == "" &&
		pos.PositionOnTI == 0
}

// IsOut is true if this position is out of the scene and moving forward
func (pos Position) IsOut() bool {
	if pos.TrackItem().Type() == TypeEnd && pos.PreviousItem() != nil {
		return true
	}
	return false
}

// Next is the first Position on the next TrackItem with regard to this Position
func (pos Position) Next(dir PointDirection) Position {
	nextTi, _ := pos.TrackItem().FollowingItem(pos.PreviousItem(), dir)
	return Position{
		simulation:     pos.simulation,
		TrackItemID:    nextTi.ID(),
		PreviousItemID: pos.TrackItemID,
		PositionOnTI:   0,
	}
}

// Previous is the last Position on the previous TrackItem with regard to this Position
func (pos Position) Previous() Position {
	previousTI, _ := pos.PreviousItem().FollowingItem(pos.TrackItem(), DirectionCurrent)
	var previousTIID string
	if previousTI != nil {
		previousTIID = previousTI.ID()
	}
	return Position{
		simulation:     pos.simulation,
		TrackItemID:    pos.PreviousItemID,
		PreviousItemID: previousTIID,
		PositionOnTI:   pos.PreviousItem().RealLength(),
	}
}

// Reversed returns the position that is at the same position but in the
// opposite direction.
func (pos Position) Reversed() Position {
	ti := pos.TrackItem()
	pti := pos.PreviousItem()
	nti, _ := ti.FollowingItem(pti, 0)
	distance := pos.TrackItem().RealLength() - pos.PositionOnTI
	return Position{
		simulation:     pos.simulation,
		TrackItemID:    ti.ID(),
		PreviousItemID: nti.ID(),
		PositionOnTI:   distance,
	}
}

// Equals returns true if this position is the same as pos2.
func (pos Position) Equals(pos2 Position) bool {
	return pos.TrackItemID == pos2.TrackItemID &&
		pos.PreviousItemID == pos2.PreviousItemID &&
		pos.PositionOnTI == pos2.PositionOnTI
}

// Add returns the Position that is length ahead of this position.
// If length is negative, find the position backwards.
func (pos Position) Add(length float64) Position {
	if length > 0 && pos.PositionOnTI+length <= pos.TrackItem().RealLength() ||
		length < 0 && pos.PositionOnTI+length >= 0 {
		return Position{
			simulation:     pos.simulation,
			TrackItemID:    pos.TrackItemID,
			PreviousItemID: pos.PreviousItemID,
			PositionOnTI:   pos.PositionOnTI + length,
		}
	}
	if length < 0 {
		return pos.Previous().Add(pos.PositionOnTI + length)
	}
	return pos.Next(DirectionCurrent).Add(length + pos.PositionOnTI - pos.TrackItem().RealLength())
}

// Sub returns the distance between orig and this position.
//
// It returns an error if both pos is not ahead of orig,
// and in the same direction
func (pos Position) Sub(orig Position) (float64, error) {
	if orig.TrackItemID == pos.TrackItemID {
		if orig.PreviousItemID != pos.PreviousItemID {
			return 0, errors.New("position is not in the same direction as orig")
		}
		if orig.PositionOnTI > pos.PositionOnTI {
			return 0, errors.New("position is not ahead of orig")
		}
		return pos.PositionOnTI - orig.PositionOnTI, nil
	}
	d, err := pos.Sub(orig.Next(DirectionCurrent))
	if err != nil {
		return 0, err
	}
	return d + orig.TrackItem().RealLength() - orig.PositionOnTI, nil
}

// trackItemsToPosition returns a list of all the trackItems between this position and
// position p including the trackItem of this position and the trackItem of position p
func (pos Position) trackItemsToPosition(p Position) []TrackItem {
	var res []TrackItem
	for cur := pos; cur.TrackItemID != p.TrackItemID && !cur.IsOut(); cur = cur.Next(DirectionCurrent) {
		res = append(res, cur.TrackItem())
	}
	res = append(res, p.TrackItem())
	return res
}

// String method for the Position type
func (pos Position) String() string {
	if pos.IsNull() {
		return "<Null Position>"
	}
	if pos.IsValid() {
		return fmt.Sprintf("(%s, %s, %0.2f)", pos.TrackItemID, pos.PreviousItemID, pos.PositionOnTI)
	}
	return fmt.Sprintf("<Invalid Position: (%v) %s, %s, %0.2f>", pos.simulation != nil, pos.TrackItemID, pos.PreviousItemID, pos.PositionOnTI)
}
