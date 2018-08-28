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

type TrackItemType string

const (
	TypeTrack         TrackItemType = "TrackItem"
	TypeLine          TrackItemType = "LineItem"
	TypeInvisibleLink TrackItemType = "InvisibleLinkItem"
	TypeEnd           TrackItemType = "EndItem"
	TypeSignal        TrackItemType = "SignalItem"
	TypePoints        TrackItemType = "PointsItem"
	TypePlace         TrackItemType = "Place"
	TypePlatform      TrackItemType = "PlatformItem"
	TypeText          TrackItemType = "TextItem"
)

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

// An ItemNotLinkedAtError is returned when a TrackItem instance has no connected item at the given end.
type ItemNotLinkedAtError struct {
	item TrackItem
	pt   Point
}

// Error method for the ItemsNotLinkedError
func (i ItemNotLinkedAtError) Error() string {
	return fmt.Sprintf("TypeTrack %s is not linked at (%f, %f)", i.item, i.pt.X, i.pt.Y)
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

	// ID returns the unique ID of this TrackItem, which is the index of this
	// item in the Simulation's TrackItems map.
	ID() int

	// Type returns the name of the type of this item
	Type() TrackItemType

	// Name returns the human readable name of this item
	Name() string

	// NextItem returns the next item of this TrackItem.
	//
	// The next item is usually the item connected to the end of the item that is not the Origin
	NextItem() TrackItem

	// PreviousItem returns the previous item of this TrackItem.
	//
	// The previous item is usually the item connected to the Origin() of this item.
	PreviousItem() TrackItem

	// MaxSpeed is the maximum allowed speed on this TrackItem in meters per second.
	MaxSpeed() float64

	// RealLength is the length in meters that this TrackItem has in real life track length
	RealLength() float64

	// Origin are the two coordinates (x, y) of the origin point of this TrackItem.
	Origin() Point

	// End are the two coordinates (xf, yf) of the end point of this TrackItem.
	End() Point

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

	// setActiveRoute sets the given route as active on this TypeTrack.
	// previous gives the direction.
	setActiveRoute(r *Route, previous TrackItem)

	// ActiveRoute returns a pointer to the route currently active on this item
	ActiveRoute() *Route

	// ActiveRoutePreviousItem returns the previous item in the active route direction
	ActiveRoutePreviousItem() TrackItem

	// underlying returns the underlying trackStruct object
	underlying() *trackStruct
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

func (t *trackStruct) ID() int {
	return t.tsId
}

// Type returns the name of the type of this item
func (t *trackStruct) Type() TrackItemType {
	return TypeTrack
}

func (t *trackStruct) Name() string {
	return t.TsName
}

func (t *trackStruct) NextItem() TrackItem {
	return t.simulation.TrackItems[t.NextTiId]
}

func (t *trackStruct) PreviousItem() TrackItem {
	return t.simulation.TrackItems[t.PreviousTiId]
}

func (t *trackStruct) MaxSpeed() float64 {
	if t.TsMaxSpeed == 0 {
		return t.simulation.Options.DefaultMaxSpeed
	}
	return t.TsMaxSpeed
}

func (t *trackStruct) RealLength() float64 {
	return t.TsRealLength
}

func (t *trackStruct) Origin() Point {
	return Point{t.X, t.Y}
}

func (t *trackStruct) End() Point {
	return Point{t.X, t.Y}
}

func (t *trackStruct) ConflictItem() TrackItem {
	return t.simulation.TrackItems[t.ConflictTiId]
}

func (t *trackStruct) Place() *Place {
	return t.simulation.Places[t.PlaceCode]
}

// FollowingItem returns the following TrackItem linked to this one,
// knowing we come from precedingItem(). Returned is either NextItem or
// PreviousItem, depending which way we come from.
//
// The second argument will return a ItemsNotLinkedError if the given
// precedingItem is not linked to this item.
func (t *trackStruct) FollowingItem(precedingItem TrackItem, dir PointDirection) (TrackItem, error) {
	if precedingItem == TrackItem(t).PreviousItem() {
		return t.NextItem(), nil
	}
	if precedingItem == TrackItem(t).NextItem() {
		return t.PreviousItem(), nil
	}
	return nil, ItemsNotLinkedError{t, precedingItem}
}

func (t *trackStruct) IsConnected(oti TrackItem) bool {
	if TrackItem(t).NextItem() == oti || TrackItem(t).PreviousItem() == t {
		return true
	}
	return false
}

func (t *trackStruct) CustomProperty(key string) CustomProperty {
	return t.CustomProperties[key]
}

func (t *trackStruct) underlying() *trackStruct {
	return t
}

// setActiveRoute sets the given route as active on this TypeTrack.
// previous gives the direction.
func (t *trackStruct) setActiveRoute(r *Route, previous TrackItem) {
	t.activeRoute = r
	t.arPreviousItem = previous
}

// ActiveRoute returns a pointer to the route currently active on this item
func (t *trackStruct) ActiveRoute() *Route {
	return t.activeRoute
}

// ActiveRoutePreviousItem returns the previous item in the active route direction
func (t *trackStruct) ActiveRoutePreviousItem() TrackItem {
	return t.arPreviousItem
}

var _ TrackItem = new(trackStruct)

// A Place is a special TrackItem representing a physical location such as a
// station or a passing point. Note that Place items are not linked to other items.
type Place struct {
	trackStruct
}

// Type returns the name of the type of this item
func (pl *Place) Type() TrackItemType {
	return TypePlace
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
func (li *LineItem) Type() TrackItemType {
	return TypeLine
}

// TrackCode returns the track number of this LineItem, if it is part of a
// TypePlace and if it has one.
func (li *LineItem) TrackCode() string {
	return li.TsTrackCode
}

// End returns the two coordinates (Xf, Yf) of the end point of this item
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
func (ili *InvisibleLinkItem) Type() TrackItemType {
	return TypeInvisibleLink
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
func (ei *EndItem) Type() TrackItemType {
	return TypeEnd
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

func (pfi *PlatformItem) Type() TrackItemType {
	return TypePlatform
}

var _ TrackItem = new(PlatformItem)

// TextItem "displays simple text" on the scenery layout
type TextItem struct {
	trackStruct
}

// Type returns the name of the type of this item
func (ti *TextItem) Type() TrackItemType {
	return TypeText
}

var _ TrackItem = new(TextItem)
