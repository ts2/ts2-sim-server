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

import "encoding/json"

// A PointsItemManager simulates the physical points, in particular delay in points
// position and breakdowns
type PointsItemManager interface {
	// Name returns a description of this PointsItemManager that is used for the UI.
	Name() string
	// Direction returns the direction of the points
	Direction(*PointsItem) PointDirection
	// SetDirection tries to set the given PointsItem to the given direction
	//
	// You should not assume that the direction has been set, since this can be
	// delayed or failed. Call Direction to check.
	SetDirection(*PointsItem, PointDirection)
}

// PointDirection are constants that represent the "physical state" of a PointsItem
type PointDirection int8

const (
	// DirectionCurrent : special position used in functions so as not to change points position
	DirectionCurrent PointDirection = -1

	// DirectionNormal : Point is set at normal
	DirectionNormal PointDirection = 0

	// DirectionReversed : Point is set for cross over
	DirectionReversed PointDirection = 1

	// DirectionUnknown : No direction is returned by the points sensors.
	// Usually means points are moving
	DirectionUnknown PointDirection = 2

	// DirectionFailed : Points or points sensors have a failure
	DirectionFailed PointDirection = 3
)

// A PointsItem is a three-way railway junction (known as Point, Switch, Turnout..)
//
// The three ends are called `common end`, `normal end` and `reverse end`
//
// 	                    ____________ reverse
// 	                   /
// 	common ___________/______________normal
//
// Trains can go from the common end to normal or reverse ends depending on the
// state of the points, but they cannot go from the normal end to reverse end.
//
// Usually, the normal end is aligned with the common end and the reverse end
// is sideways, but this is not mandatory.
//
// Geometric points are represented on a 10 x 10 square centered on Center() point. CommonEnd,
// NormalEnd and ReverseEnd are points on the side of this square (i.e. they have
// at least one coordinate which is 5 or -5)
type PointsItem struct {
	trackStruct
	Xc          float64 `json:"xf"`
	Yc          float64 `json:"yf"`
	Xn          float64 `json:"xn"`
	Yn          float64 `json:"yn"`
	Xr          float64 `json:"xr"`
	Yr          float64 `json:"yr"`
	ReverseTiId string  `json:"reverseTiId"`
	PairedTiId  string  `json:"pairedTiId"`
}

// Type returns the name of the type of this item
func (pi *PointsItem) Type() TrackItemType {
	return TypePoints
}

// Origin are the two coordinates (x, y) of the origin, i.e. the absolute coordinates of the common end.
func (pi *PointsItem) Origin() Point {
	return pi.Center().Add(pi.CommonEnd())
}

// End returns the two coordinates (Xf, Yf) of the end, i.e. the absolute coordinates of the normal end.
func (pi *PointsItem) End() Point {
	return pi.Center().Add(pi.NormalEnd())
}

// Reverse returns the two (Xr, Yr) absolute coordinates of the reverse end.
func (pi *PointsItem) Reverse() Point {
	return pi.Center().Add(pi.ReverseEnd())
}

// Center point of this PointsItem in the scene coordinates
func (pi *PointsItem) Center() Point {
	return Point{pi.X, pi.Y}
}

// CommonEnd return the common end point in the item's coordinates
func (pi *PointsItem) CommonEnd() Point {
	return Point{pi.Xc, pi.Yc}
}

// NormalEnd return the normal end point in the item's coordinates
func (pi *PointsItem) NormalEnd() Point {
	return Point{pi.Xn, pi.Yn}
}

// ReverseEnd return the reverse end point in the item's coordinates
func (pi *PointsItem) ReverseEnd() Point {
	return Point{pi.Xr, pi.Yr}
}

// ReverseItem returns the item linked to the reverse end of these points
func (pi *PointsItem) ReverseItem() TrackItem {
	return pi.simulation.TrackItems[pi.ReverseTiId]
}

// PairedItem returns the points item that must change simultaneously with
// this one. It return nil if there is no such item.
func (pi *PointsItem) PairedItem() *PointsItem {
	paired, ok := pi.simulation.TrackItems[pi.PairedTiId].(*PointsItem)
	if ok {
		return paired
	}
	return nil
}

// Reversed returns true if the points are in the reversed position, false
// otherwise
func (pi *PointsItem) Reversed() bool {
	dir := pointsItemManager.Direction(pi)
	return dir == DirectionReversed
}

// IsConnected returns true if this TrackItem is connected to the given
// TrackItem, false otherwise
func (pi *PointsItem) IsConnected(oti TrackItem) bool {
	if pi.trackStruct.IsConnected(oti) {
		return true
	}
	if pi.ReverseTiId == oti.ID() {
		return true
	}
	return false
}

// FollowingItem returns the following TrackItem linked to this one,
// knowing we come from precedingItem(). Returned is either NextItem or
// PreviousItem, depending which way we come from.
//
// The second argument will return a ItemsNotLinkedError if the given
// precedingItem is not linked to this item.
func (pi *PointsItem) FollowingItem(precedingItem TrackItem, dir PointDirection) (TrackItem, error) {
	if precedingItem == pi.ReverseItem() || precedingItem == pi.NextItem() {
		return pi.PreviousItem(), nil
	}
	if precedingItem == pi.PreviousItem() {
		switch dir {
		case DirectionReversed:
			return pi.ReverseItem(), nil
		case DirectionNormal, DirectionFailed, DirectionUnknown:
			return pi.NextItem(), nil
		case DirectionCurrent:
			if pi.Reversed() {
				return pi.ReverseItem(), nil
			}
			return pi.NextItem(), nil
		}
	}
	return nil, ItemsNotLinkedError{pi, precedingItem}
}

// setActiveRoute sets the given route as active on this PointsItem.
// previous gives the direction.
func (pi *PointsItem) setActiveRoute(r *Route, previous TrackItem) {
	if r != nil {
		pointsItemManager.SetDirection(pi, r.Directions[pi.ID()])
	}
	// Send event for pairedItem
	if pi.PairedItem() != nil {
		pi.simulation.sendEvent(&Event{
			Name:   TrackItemChangedEvent,
			Object: pi.PairedItem(),
		})
	}
	// TODO We should check here whether the points have failed or not
	// and delay route activation.
	pi.trackStruct.setActiveRoute(r, previous)
}

// MarshalJSON method for PointsItem
func (pi *PointsItem) MarshalJSON() ([]byte, error) {
	type auxPI struct {
		jsonTrackStruct
		Xc          float64 `json:"xf"`
		Yc          float64 `json:"yf"`
		Xn          float64 `json:"xn"`
		Yn          float64 `json:"yn"`
		Xr          float64 `json:"xr"`
		Yr          float64 `json:"yr"`
		ReverseTiId string  `json:"reverseTiId"`
		PairedTiId  string  `json:"pairedTiId"`
		Reversed    bool    `json:"reversed"`
	}
	aPI := auxPI{
		jsonTrackStruct: pi.asJSONStruct(),
		Xc:              pi.Xc,
		Yc:              pi.Yc,
		Xn:              pi.Xn,
		Yn:              pi.Yn,
		Xr:              pi.Xr,
		Yr:              pi.Yr,
		ReverseTiId:     pi.ReverseTiId,
		PairedTiId:      pi.PairedTiId,
		Reversed:        pi.Reversed(),
	}
	return json.Marshal(aPI)
}

var _ TrackItem = new(PointsItem)
