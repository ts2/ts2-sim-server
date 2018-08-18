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
	"fmt"
)

// bigFloat is a large number used for the length of an EndItem. It must be bigger
// than the maximum distance the fastest train can travel during the game time step
// at maximum simulation speed.
const bigFloat = 1000000000.0

type CustomProperty map[string][]int

// An ItemsNotLinkedError is returned when two TrackItem instances that are assumed
// to be linked are not.
type ItemsNotLinkedError struct {
	item1 TrackItem
	item2 TrackItem
}

// Error method for the ItemsNotLinkedError
func (e ItemsNotLinkedError) Error() string {
	return fmt.Sprintf("TrackItems %s and %s are not linked", e.item1, e.item2)
}

// A TrackItem is a piece of scenery and is "the base interface" for others
// such as SignalItem, EndItem, PointsItem.
//
// Every item has defined coordinates in the scenery layout and is connected to other
// TrackItems's so that trains can travel from one to another.
//
// The coordinates are expressed in pixels, the X-axis is from left to right and
// the Y-axis is from top to bottom.
//
// Every TrackItem has an Origin() Point defined by its X and Y values.
type TrackItem interface {

	// TiID returns the unique ID of this TrackItem, which is the index of this
	// item in the Simulation's TrackItems map.
	TiID() int

	// Type returns the name of the type of this item
	Type() string

	// Name returns the human readable name of this item
	Name() string

	// setSimulation attaches this TrackItem to a Simulation instance
	setSimulation(*Simulation)

	// setID() sets the item's internal id
	setID(int)

	// NextItem() returns the next item of this TrackItem.
	//
	// The next item is usually the item connected to the end of the item that is not the Origin()
	NextItem() TrackItem

	// PreviousItem returns the previous item of this TrackItem.
	//
	// The previous item is usually the item connected to the Origin() of this item.
	PreviousItem() TrackItem

	// MaxSpeed() is the maximum allowed speed on this TrackItem in meters per second.
	MaxSpeed() float64

	// RealLength() is the length in meters that this TrackItem has in real life track length
	RealLength() float64

	// Origin() are the two coordinates (x, y) of the origin point of this TrackItem.
	Origin() Point

	// ConflictItem returns the conflicting item of this TrackItem. The conflicting
	// item is another item of the scenery on which a route must not be set if
	// one is already active on this TrackItem (and vice-versa). This is
	// particularly the case when two TrackItems cross over with no points.
	ConflictItem() TrackItem

	// Place returns the TrackItem of type Place associated with this item
	// (as defined by PlaceCode).
	Place() *Place

	// FollowingItem returns the following TrackItem linked to this one,
	// knowing we come from precedingItem(). Returned is either NextItem or
	// PreviousItem, depending which way we come from.
	//
	// The second argument will return a ItemsNotLinkedError if the given
	// precedingItem is not linked to this item.
	FollowingItem(TrackItem, PointDirection) (TrackItem, error)

	// IsConnected returns true if this TrackItem is connected to the given
	// TrackItem, false otherwise
	IsConnected(TrackItem) bool

	// CustomProperty returns the custom property with the given key
	CustomProperty(string) CustomProperty
}

// trackStruct is an abstract struct the pointer of which implements TrackItem
type trackStruct struct {
	TiType           string                    `json:"__type__"`
	TsName           string                    `json:"name"`
	NextTiId         int                       `json:"nextTiId"`
	PreviousTiId     int                       `json:"previousTiId"`
	TsMaxSpeed       float64                   `json:"maxSpeed"`
	TsRealLength     float64                   `json:"realLength"`
	X                float64                   `json:"x"`
	Y                float64                   `json:"y"`
	ConflictTiId     int                       `json:"conflictTiId"`
	CustomProperties map[string]CustomProperty `json:"customProperties"`
	PlaceCode        string                    `json:"placeCode"`

	tsId           int
	simulation     *Simulation
	activeRoute    *Route
	arPreviousItem TrackItem
	selected       bool
	trains         []*Train
}

func (ti *trackStruct) TiID() int {
	return ti.tsId
}

// Type returns the name of the type of this item
func (ti *trackStruct) Type() string {
	return "TrackItem"
}

func (ti *trackStruct) Name() string {
	return ti.TsName
}

func (ti *trackStruct) setSimulation(sim *Simulation) {
	ti.simulation = sim
}

func (ti *trackStruct) setID(tiId int) {
	ti.tsId = tiId
}

func (ti *trackStruct) NextItem() TrackItem {
	return ti.simulation.TrackItems[ti.NextTiId]
}

func (ti *trackStruct) PreviousItem() TrackItem {
	return ti.simulation.TrackItems[ti.PreviousTiId]
}

func (ti *trackStruct) MaxSpeed() float64 {
	if ti.TsMaxSpeed == 0 {
		return ti.simulation.Options.DefaultMaxSpeed
	}
	return ti.TsMaxSpeed
}

func (ti *trackStruct) RealLength() float64 {
	return ti.TsRealLength
}

func (ti *trackStruct) Origin() Point {
	return Point{ti.X, ti.Y}
}

func (ti *trackStruct) ConflictItem() TrackItem {
	return ti.simulation.TrackItems[ti.ConflictTiId]
}

func (ti *trackStruct) Place() *Place {
	return ti.simulation.Places[ti.PlaceCode]
}

// FollowingItem returns the following TrackItem linked to this one,
// knowing we come from precedingItem(). Returned is either NextItem or
// PreviousItem, depending which way we come from.
//
// The second argument will return a ItemsNotLinkedError if the given
// precedingItem is not linked to this item.
func (ti *trackStruct) FollowingItem(precedingItem TrackItem, dir PointDirection) (TrackItem, error) {
	if precedingItem == TrackItem(ti).PreviousItem() {
		return ti.NextItem(), nil
	}
	if precedingItem == TrackItem(ti).NextItem() {
		return ti.PreviousItem(), nil
	}
	return nil, ItemsNotLinkedError{ti, precedingItem}
}

func (ti *trackStruct) IsConnected(oti TrackItem) bool {
	if TrackItem(ti).NextItem() == oti || TrackItem(ti).PreviousItem() == ti {
		return true
	}
	return false
}

func (ti *trackStruct) CustomProperty(key string) CustomProperty {
	return ti.CustomProperties[key]
}

var _ TrackItem = new(trackStruct)

// ResizableItem is the base of any TrackItem that can be resized by the user in
// the editor, such as LineItem or PlatformItem.
type ResizableItem interface {
	TrackItem
	// End returns the two coordinates (Xf, Yf) of the end point of this
	// ResizeableItem.
	End() Point
}

// A Place is a special TrackItem representing a physical location such as a
// station or a passing point. Note that Place items are not linked to other items.
type Place struct {
	trackStruct
}

// Type returns the name of the type of this item
func (pl *Place) Type() string {
	return "Place"
}

var _ TrackItem = new(Place)

// A LineItem is a resizable TrackItem that represent a simple railway line and
// is used to connect two TrackItem's together.
type LineItem struct {
	trackStruct
	Xf          float64 `json:"xf"`
	Yf          float64 `json:"yf"`
	TsTrackCode string  `json:"trackCode"`
}

// Type returns the name of the type of this item
func (li *LineItem) Type() string {
	return "LineItem"
}

// TrackCode returns the track number of this LineItem, if it is part of a
// place and if it has one.
func (li *LineItem) TrackCode() string {
	return li.TsTrackCode
}

// End returns the two coordinates (Xf, Yf) of the end point of this
// ResizeableItem.
func (li *LineItem) End() Point {
	return Point{li.Xf, li.Yf}
}

var _ TrackItem = new(LineItem)

// InvisibleLinkItem behave like line items, but clients are encouraged not to
// represented them on the scenery. They are used to make links between lines or to
// represent bridges and tunnels.
type InvisibleLinkItem struct {
	LineItem
}

// Type returns the name of the type of this item
func (ili *InvisibleLinkItem) Type() string {
	return "InvisibleLinkItem"
}

var _ TrackItem = new(InvisibleLinkItem)

// An EndItem is an invisible item to which the free ends of other Trackitem instances
// must be connected to prevent the simulation from crashing.
//
// End items are single point items.
type EndItem struct {
	trackStruct
}

// Type returns the name of the type of this item
func (ei *EndItem) Type() string {
	return "EndItem"
}

// RealLength() is the length in meters that this TrackItem has in real life track length
func (ei *EndItem) RealLength() float64 {
	return bigFloat
}

var _ TrackItem = new(EndItem)

// PlatformItem's are usually represented as a colored rectangle on the scene to
// symbolise the platform. This colored rectangle can permit user interaction.
type PlatformItem struct {
	LineItem
}

func (pfi *PlatformItem) Type() string {
	return "PlatformItem"
}

var _ TrackItem = new(PlatformItem)

// TextItem "displays simple text" on the scenery layout
type TextItem struct {
	trackStruct
}

// Type returns the name of the type of this item
func (ti *TextItem) Type() string {
	return "TextItem"
}

var _ TrackItem = new(TextItem)
