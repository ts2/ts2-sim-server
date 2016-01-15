/*   Copyright (C) 2008-2016 by Nicolas Piganeau and the TS2 TEAM
 *   (See AUTHORS file)
 *
 *   This program is free software; you can redistribute it and/or modify
 *   it under the terms of the GNU General Public License as published by
 *   the Free Software Foundation; either version 2 of the License, or
 *   (at your option) any later version.
 *
 *   This program is distributed in the hope that it will be useful,
 *   but WITHOUT ANY WARRANTY; without even the implied warranty of
 *   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *   GNU General Public License for more details.
 *
 *   You should have received a copy of the GNU General Public License
 *   along with this program; if not, write to the
 *   Free Software Foundation, Inc.,
 *   59 Temple Place - Suite 330, Boston, MA  02111-1307, USA.
 */

package simulation

import (
	//"fmt"
)


// PointDirection are constants that represent the "physical state" of a PointsItem
type PointDirection uint8

const (
// Rail change is set at normal
	NORMAL PointDirection = 0

// Rail change is set for cross over
	REVERSED PointDirection = 1

// Point is moving and Unknown state
	MOVING PointDirection = 10

// Point goes back to previous safe state.. and fail
	BACKOFF PointDirection = 11
)




/*
A `PointsItem` is a three-way railway junction (known as Point, Switch, Turnout..)

The three ends are called `common end`, `normal end` and `reverse end`

	                    ____________ reverse
	                   /
	common ___________/______________normal

Trains can go from the common end to normal or reverse ends depending on the
state of the points, but they cannot go from the normal end to reverse end.

Usually, the normal end is aligned with the common end and the reverse end
is sideways, but this is not mandatory.

Geometric points are represented on a 10 x 10 square centered on Center() point. CommonEnd,
NormalEnd and ReverseEnd are points on the side of this square (i.e. they have
at least one coordinate which is 5 or -5)
*/
type PointsItem interface {
	TrackItem

	// The center point of this PointsItem in the scene coordinates
	Center() Point

	// CommonEnd return the common end point in the item's coordinates
	CommonEnd() Point

	// NormalEnd return the normal end point in the item's coordinates
	NormalEnd() Point

	// ReverseEnd return the reverse end point in the item's coordinates
	ReverseEnd() Point

	// ReversedItem returns the item linked to the reverse end of these points
	ReverseItem() TrackItem

	// Reversed returns true if the points are in the reversed position, false
	// otherwise
	Reversed() bool
}

/*
pointsStruct is a struct which implements PointsItem
*/
type pointsStruct struct {
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

func (pi *pointsStruct) Type() string {
	return "PointsItem"
}

func (pi *pointsStruct) Center() Point {
	return Point{pi.X, pi.Y}
}

func (pi *pointsStruct) CommonEnd() Point {
	return Point{pi.Xc, pi.Yc}
}

func (pi *pointsStruct) NormalEnd() Point {
	return Point{pi.Xn, pi.Yn}
}

func (pi *pointsStruct) ReverseEnd() Point {
	return Point{pi.Xr, pi.Yr}
}

func (pi *pointsStruct) ReverseItem() TrackItem {
	return pi.simulation.TrackItems[pi.ReverseTiId]
}

func (pi *pointsStruct) Reversed() bool {
	return pi.reversed
}
func (ti *pointsStruct) FollowingItem(precedingItem TrackItem, dir PointDirection) (TrackItem, error) {
	if precedingItem == PointsItem(ti).ReverseItem() || precedingItem == PointsItem(ti).NextItem() {
		return ti.PreviousItem(), nil
	}
	if precedingItem == PointsItem(ti).PreviousItem() {
		if dir == REVERSED {
			return ti.ReverseItem(), nil
		} else {
			return ti.NextItem(), nil
		}
	}
	return nil, ItemsNotLinkedError{ti, precedingItem}
}

