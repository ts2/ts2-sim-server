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

// PointDirection are constants that represent the "physical state" of a PointsItem
type PointDirection uint8

const (
	// normal : Point is set at normal
	normal PointDirection = 0

	// reversed : Point is set for cross over
	reversed PointDirection = 1
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
	ReverseTiId int     `json:"reverseTiId"`
	reversed    bool
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

// Reversed returns true if the points are in the reversed position, false
// otherwise
func (pi *PointsItem) Reversed() bool {
	return pi.reversed
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
		if dir == reversed {
			return pi.ReverseItem(), nil
		} else {
			return pi.NextItem(), nil
		}
	}
	return nil, ItemsNotLinkedError{pi, precedingItem}
}

// setActiveRoute sets the given route as active on this PointsItem.
// previous gives the direction.
func (pi *PointsItem) setActiveRoute(r *Route, previous TrackItem) {
	pi.reversed = false
	if r.Directions[pi.ID()] != 0 {
		pi.reversed = true
	}
	pi.trackStruct.setActiveRoute(r, previous)
}

var _ TrackItem = new(PointsItem)
