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
	"encoding/json"
	"fmt"
	"sync"
)

// bigFloat is a large number used for the length of an EndItem. It must be bigger
// than the maximum distance the fastest train can travel during the game time step
// at maximum simulation speed.
const bigFloat = 1000000000.0

// NoMorePlace is a large float used to represent a non existent service line index
const NoMorePlace = 9999

// A TrackItemType holds the type of a track item
type TrackItemType string

// Available track item types.
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

// A CustomProperty is a map to hold track item properties that are defined by the user.
type CustomProperty map[string][]string

// A LineItemManager manages breakdowns of line track circuits
type LineItemManager interface {
	// Name returns a description of this lineItemManager that is used for the UI.
	Name() string
	// IsFailed returns true if the given LineItem has a track circuit failure
	IsFailed(*LineItem) bool
}

// An ItemsNotLinkedError is returned when two TrackItem instances that are assumed
// to be linked are not.
type ItemsNotLinkedError struct {
	item1 TrackItem
	item2 TrackItem
}

// Error method for the ItemsNotLinkedError
func (e ItemsNotLinkedError) Error() string {
	return fmt.Sprintf("TrackItems %s and %s are not linked", e.item1.ID(), e.item2.ID())
}

// An ItemNotLinkedAtError is returned when a TrackItem instance has no connected item at the given end.
type ItemNotLinkedAtError struct {
	item TrackItem
	pt   Point
}

// Error method for the ItemsNotLinkedError
func (i ItemNotLinkedAtError) Error() string {
	return fmt.Sprintf("TrackItem %s is not linked at (%f, %f)", i.item.ID(), i.pt.X, i.pt.Y)
}

// ItemInconsistentLinkError is returned when a TrackItem is linked to another one, but
// the latter is not linked to the former.
type ItemInconsistentLinkError struct {
	item1 TrackItem
	item2 TrackItem
	pt    Point
}

// Error method for the ItemInconsistentLinkError
func (i ItemInconsistentLinkError) Error() string {
	return fmt.Sprintf("inconsistent link at (%f, %f) between %s and %s", i.pt.X, i.pt.Y, i.item1.ID(), i.item2.ID())
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
	// routeID returns the unique routeID of this TrackItem, which is the index of this
	// item in the Simulation's TrackItems map.
	ID() string

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

	// TrackCode returns the code (usually a number) of the track line
	TrackCode() string

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

	// trainHeadActions performs the actions to be done when a train head reaches this TrackItem
	trainHeadActions(*Train)

	// trainTailActions performs the actions to be done when a train tail reaches this TrackItem
	trainTailActions(*Train)

	// resetActiveRoute resets route information on this item.
	resetActiveRoute()

	// notifyChange sends a TrackItemChanged event for this item
	notifyChange()

	// TrainPresent returns true if at least one train is present on this TrackItem
	TrainPresent() bool

	// IsOnPosition returns true if this track item is the track item of the given position.
	// When applicable, also checks if the item is in the same direction as the position.
	IsOnPosition(Position) bool

	// DistanceToTrainEnd returns the distance to the closest end (either train head or
	// train tail) of the closest train when on pos. If no train is on this item, the
	// distance will be 0, and the second argument will be false.
	DistanceToTrainEnd(Position) (float64, bool)

	// Equals returns true if this track item and the given one are the same
	// (i.e. they have the same routeID)
	Equals(TrackItem) bool

	// addTrigger adds the given function to the list of functions that will be
	// called when a trains enters this TrackItem.
	addTrigger(func(TrackItem))

	// Simulation returns the Simulation object that this TrackItem belongs to.
	Simulation() *Simulation

	// initialize this TrackItem
	initialize() error

	// underlying returns the underlying trackStruct object
	underlying() *trackStruct
}

// trackStruct is an abstract struct the pointer of which implements TrackItem
type trackStruct struct {
	TiType           string                    `json:"__type__"`
	TsName           string                    `json:"name"`
	NextTiID         string                    `json:"nextTiId"`
	PreviousTiID     string                    `json:"previousTiId"`
	TsMaxSpeed       float64                   `json:"maxSpeed"`
	TsRealLength     float64                   `json:"realLength"`
	X                float64                   `json:"x"`
	Y                float64                   `json:"y"`
	ConflictTiId     string                    `json:"conflictTiId"`
	CustomProperties map[string]CustomProperty `json:"customProperties"`
	PlaceCode        string                    `json:"placeCode"`
	TsTrackCode      string                    `json:"trackCode"`

	tsId           string
	simulation     *Simulation
	activeRoute    *Route
	arPreviousItem TrackItem
	selected       bool
	trainEndsFW    map[*Train]float64
	trainEndsBK    map[*Train]float64
	trainEndMutex  sync.RWMutex
	triggers       []func(TrackItem)
}

// routeID returns the unique routeID of this TrackItem, which is the index of this
// item in the Simulation's TrackItems map.
func (t *trackStruct) ID() string {
	return t.tsId
}

// Type returns the name of the type of this item
func (t *trackStruct) Type() TrackItemType {
	return TypeTrack
}

// Name returns the human readable name of this item
func (t *trackStruct) Name() string {
	return t.TsName
}

// NextItem returns the next item of this TrackItem.
//
// The next item is usually the item connected to the end of the item that is not the Origin
func (t *trackStruct) NextItem() TrackItem {
	return t.simulation.TrackItems[t.NextTiID]
}

// PreviousItem returns the previous item of this TrackItem.
//
// The previous item is usually the item connected to the Origin() of this item.
func (t *trackStruct) PreviousItem() TrackItem {
	return t.simulation.TrackItems[t.PreviousTiID]
}

// MaxSpeed is the maximum allowed speed on this TrackItem in meters per second.
func (t *trackStruct) MaxSpeed() float64 {
	if t.TsMaxSpeed == 0 {
		return t.simulation.Options.DefaultMaxSpeed
	}
	return t.TsMaxSpeed
}

// RealLength is the length in meters that this TrackItem has in real life track length
func (t *trackStruct) RealLength() float64 {
	return t.TsRealLength
}

// Origin are the two coordinates (x, y) of the origin point of this TrackItem.
func (t *trackStruct) Origin() Point {
	return Point{t.X, t.Y}
}

// End are the two coordinates (xf, yf) of the end point of this TrackItem.
func (t *trackStruct) End() Point {
	return Point{t.X, t.Y}
}

// ConflictItem returns the conflicting item of this TrackItem. The conflicting
// item is another item of the scenery on which a route must not be set if
// one is already active on this TrackItem (and vice-versa). This is
// particularly the case when two TrackItems cross over with no points.
func (t *trackStruct) ConflictItem() TrackItem {
	return t.simulation.TrackItems[t.ConflictTiId]
}

// Place returns the TrackItem of type Place associated with this item
// (as defined by PlaceCode).
func (t *trackStruct) Place() *Place {
	return t.simulation.Places[t.PlaceCode]
}

// TrackCode returns the track number of this LineItem, if it is part of a
// TypePlace and if it has one.
func (t *trackStruct) TrackCode() string {
	return t.TsTrackCode
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

// IsConnected returns true if this TrackItem is connected to the given
// TrackItem, false otherwise
func (t *trackStruct) IsConnected(oti TrackItem) bool {
	if t.NextTiID == oti.ID() || t.PreviousTiID == oti.ID() {
		return true
	}
	return false
}

// CustomProperty returns the custom property with the given key
func (t *trackStruct) CustomProperty(key string) CustomProperty {
	return t.CustomProperties[key]
}

// addTrigger adds the given function to the list of functions that will be
// called when a trains enters this TrackItem.
func (t *trackStruct) addTrigger(trigger func(TrackItem)) {
	t.triggers = append(t.triggers, trigger)
}

// Simulation returns the Simulation object that this TrackItem belongs to.
func (t *trackStruct) Simulation() *Simulation {
	return t.simulation
}

func (t *trackStruct) underlying() *trackStruct {
	return t
}

// setActiveRoute sets the given route as active on this TypeTrack.
// previous gives the direction.
func (t *trackStruct) setActiveRoute(r *Route, previous TrackItem) {
	t.activeRoute = r
	t.arPreviousItem = previous
	t.notifyChange()
}

// ActiveRoute returns a pointer to the route currently active on this item
func (t *trackStruct) ActiveRoute() *Route {
	return t.activeRoute
}

// ActiveRoutePreviousItem returns the previous item in the active route direction
func (t *trackStruct) ActiveRoutePreviousItem() TrackItem {
	return t.arPreviousItem
}

// trainHeadActions performs the actions to be done when a train head reaches this TrackItem
func (t *trackStruct) trainHeadActions(train *Train) {
	for _, trigger := range t.triggers {
		trigger(t)
	}
}

// trainTailActions performs the actions to be done when a train tail reaches this TrackItem
func (t *trackStruct) trainTailActions(train *Train) {
	for _, trigger := range t.triggers {
		trigger(t)
	}
	t.releaseRouteBehind()
}

// releaseRouteBehind automatically releases the route after train passed if applicable
func (t *trackStruct) releaseRouteBehind() {
	if t.activeRoute == nil {
		return
	}
	if t.activeRoute.State() != Activated && t.activeRoute.State() != Destroying {
		return
	}
	beginSignalNextRoute := t.activeRoute.BeginSignal().nextActiveRoute
	if beginSignalNextRoute != nil && beginSignalNextRoute.routeID == t.activeRoute.routeID {
		// same route has been set again
		return
	}
	if t.ActiveRoutePreviousItem().ActiveRoute() == nil || !t.ActiveRoutePreviousItem().ActiveRoute().Equals(t.activeRoute) {
		// previous item has been already set to a new route which is not ours
		return
	}
	t.ActiveRoutePreviousItem().resetActiveRoute()
}

// TrainPresent returns true if at least one train is present on this TrackItem
func (t *trackStruct) TrainPresent() bool {
	t.trainEndMutex.RLock()
	defer t.trainEndMutex.RUnlock()
	return len(t.trainEndsFW)+len(t.trainEndsBK) > 0
}

// resetActiveRoute resets route information on this item.
func (t *trackStruct) resetActiveRoute() {
	t.activeRoute = nil
	t.arPreviousItem = nil
	t.notifyChange()
}

// notifyChange sends a TrackItemChangedEvent for this item
func (t *trackStruct) notifyChange() {
	t.simulation.sendEvent(&Event{
		Name:   TrackItemChangedEvent,
		Object: t.full(),
	})
}

// IsOnPosition returns true if this track item is the track item of the given position.
func (t *trackStruct) IsOnPosition(pos Position) bool {
	return pos.TrackItemID == t.ID()
}

// DistanceToTrainEnd returns the distance to the closest end (either train head or
// train tail) of the closest train when on pos. If no train is on this item, the
// distance will be 0, and the second argument will be false.
func (t *trackStruct) DistanceToTrainEnd(pos Position) (float64, bool) {
	t.trainEndMutex.RLock()
	defer t.trainEndMutex.RUnlock()
	var mdSet bool
	minDist := bigFloat
	if pos.PreviousItemID == t.PreviousTiID {
		for _, teb := range t.trainEndsBK {
			x := teb - pos.PositionOnTI
			if x > 0 && x < minDist {
				minDist = x
				mdSet = true
			}
		}
		if !mdSet {
			return 0, false
		}
		return minDist, true
	}
	for _, tef := range t.trainEndsFW {
		x := t.RealLength() - tef - pos.PositionOnTI
		if x > 0 && x < minDist {
			minDist = x
			mdSet = true
		}
		if !mdSet {
			return 0, false
		}
		return minDist, true
	}
	return minDist, mdSet
}

// Equals returns true if this track item and the given one are the same
// (i.e. they have the same routeID)
func (t *trackStruct) Equals(ti TrackItem) bool {
	if ti == nil {
		return false
	}
	return t.ID() == ti.ID()
}

// initialize this track item
func (t *trackStruct) initialize() error {
	t.trainEndMutex.Lock()
	defer t.trainEndMutex.Unlock()
	t.trainEndsFW = make(map[*Train]float64)
	t.trainEndsBK = make(map[*Train]float64)
	return nil
}

// full returns this item as a TrackItem by reloading it
func (t *trackStruct) full() TrackItem {
	return t.simulation.TrackItems[t.ID()]
}

// MarshalJSON method for trackStruct
func (t *trackStruct) MarshalJSON() ([]byte, error) {
	ai := t.asJSONStruct()
	return json.Marshal(ai)
}

// asJSONStruct returns this trackStruct as a jsonTrackStruct
func (t *trackStruct) asJSONStruct() jsonTrackStruct {
	var arID, arpiID string
	if t.activeRoute != nil {
		arID = t.activeRoute.ID()
	}
	if t.arPreviousItem != nil {
		arpiID = t.arPreviousItem.ID()
	}
	t.trainEndMutex.RLock()
	defer t.trainEndMutex.RUnlock()
	tEndsFW := make(map[string]float64)
	for t, p := range t.trainEndsFW {
		tEndsFW[t.ID()] = p
	}
	tEndsBK := make(map[string]float64)
	for t, p := range t.trainEndsBK {
		tEndsBK[t.ID()] = p
	}
	ai := jsonTrackStruct{
		ID:               t.ID(),
		TiType:           t.TiType,
		TsName:           t.TsName,
		NextTiID:         t.NextTiID,
		PreviousTiID:     t.PreviousTiID,
		TsMaxSpeed:       t.TsMaxSpeed,
		TsRealLength:     t.TsRealLength,
		X:                t.X,
		Y:                t.Y,
		ConflictTiId:     t.ConflictTiId,
		CustomProperties: t.CustomProperties,
		PlaceCode:        t.PlaceCode,
		ActiveRoute:      arID,
		ARPreviousItem:   arpiID,
		TrainEndsFW:      tEndsFW,
		TrainEndsBK:      tEndsBK,
		TsTrackCode:      t.TsTrackCode,
	}
	return ai
}

var _ TrackItem = new(trackStruct)

type jsonTrackStruct struct {
	ID               string                    `json:"id"`
	TiType           string                    `json:"__type__"`
	TsName           string                    `json:"name"`
	NextTiID         string                    `json:"nextTiId"`
	PreviousTiID     string                    `json:"previousTiId"`
	TsMaxSpeed       float64                   `json:"maxSpeed"`
	TsRealLength     float64                   `json:"realLength"`
	X                float64                   `json:"x"`
	Y                float64                   `json:"y"`
	ConflictTiId     string                    `json:"conflictTiId"`
	CustomProperties map[string]CustomProperty `json:"customProperties"`
	PlaceCode        string                    `json:"placeCode"`
	ActiveRoute      string                    `json:"activeRoute"`
	ARPreviousItem   string                    `json:"activeRoutePreviousItem"`
	TrainEndsFW      map[string]float64        `json:"trainEndsFW"`
	TrainEndsBK      map[string]float64        `json:"trainEndsBK"`
	TsTrackCode      string                    `json:"trackCode"`
}

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
	Xf float64 `json:"xf"`
	Yf float64 `json:"yf"`
}

// Type returns the name of the type of this item
func (li *LineItem) Type() TrackItemType {
	return TypeLine
}

// End returns the two coordinates (Xf, Yf) of the end point of this item
func (li *LineItem) End() Point {
	return Point{li.Xf, li.Yf}
}

// MarshalJSON method for LineItem
func (li *LineItem) MarshalJSON() ([]byte, error) {
	type auxLI struct {
		jsonTrackStruct
		Xf float64 `json:"xf"`
		Yf float64 `json:"yf"`
	}
	aLI := auxLI{
		jsonTrackStruct: li.asJSONStruct(),
		Xf:              li.Xf,
		Yf:              li.Yf,
	}
	return json.Marshal(aLI)
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

// MarshalJSON method for the end item
func (ei *EndItem) MarshalJSON() ([]byte, error) {
	return json.Marshal(ei.asJSONStruct())
}

var _ TrackItem = new(EndItem)

// PlatformItem's are usually represented as a colored rectangle on the scene to
// symbolise the platform. This colored rectangle can permit user interaction.
type PlatformItem struct {
	LineItem
}

// Type returns the name of the type of this item
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
