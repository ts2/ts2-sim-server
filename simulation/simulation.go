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
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	log "gopkg.in/inconshreveable/log15.v2"
)

const timeStep = 500 * time.Millisecond

var (
	logger               log.Logger
	routesManagers       []RoutesManager
	trainsManagers       map[string]TrainsManager
	lineItemManager      LineItemManager
	pointsItemManager    PointsItemManager
	signalItemManager    SignalItemManager
	defaultTrainManager  TrainsManager
	signalConditionTypes map[string]ConditionType
)

// InitializeLogger creates the logger for the simulation module
func InitializeLogger(parentLogger log.Logger) {
	logger = parentLogger.New("module", "simulation")
}

// Simulation holds all the game logic.
type Simulation struct {
	SignalLib     SignalLibrary
	TrackItems    map[string]TrackItem
	Places        map[string]*Place
	Options       Options
	Routes        map[string]*Route
	TrainTypes    map[string]*TrainType
	Services      map[string]*Service
	Trains        []*Train
	MessageLogger *MessageLogger
	EventChan     chan *Event

	clockTicker *time.Ticker
	stopChan    chan bool
	started     bool
}

// UnmarshalJSON for the Simulation type
func (sim *Simulation) UnmarshalJSON(data []byte) error {
	type auxItem map[string]json.RawMessage

	type auxSim struct {
		TrackItems    map[string]json.RawMessage
		Options       Options
		SignalLib     SignalLibrary         `json:"signalLibrary"`
		Routes        map[string]*Route     `json:"routes"`
		TrainTypes    map[string]*TrainType `json:"trainTypes"`
		Services      map[string]*Service   `json:"services"`
		Trains        []*Train              `json:"trains"`
		MessageLogger *MessageLogger        `json:"messageLogger"`
	}

	var rawSim auxSim
	if err := json.Unmarshal(data, &rawSim); err != nil {
		return fmt.Errorf("unable to decode simulation JSON: %s", err)
	}
	sim.SignalLib = rawSim.SignalLib
	if err := sim.SignalLib.initialize(); err != nil {
		logger.Crit("error initializing signal Library", "error", err)
		panic(err)
	}
	sim.TrackItems = make(map[string]TrackItem)
	sim.Places = make(map[string]*Place)
	for tiId, tiString := range rawSim.TrackItems {
		var rawItem auxItem
		if err := json.Unmarshal(tiString, &rawItem); err != nil {
			return fmt.Errorf("unable to read TrackItem: %s. %s", tiString, err)
		}

		tiType := string(rawItem["__type__"])
		unmarshalItem := func(ti TrackItem) error {
			if err := json.Unmarshal(tiString, ti); err != nil {
				return fmt.Errorf("unable to decode %s: %s. %s", tiType, tiString, err)
			}
			ti.underlying().simulation = sim
			ti.underlying().tsId = tiId
			if err := ti.initialize(); err != nil {
				return err
			}
			sim.TrackItems[tiId] = ti
			return nil
		}

		switch tiType {
		case `"LineItem"`:
			var ti LineItem
			unmarshalItem(&ti)
		case `"InvisibleLinkItem"`:
			var ti InvisibleLinkItem
			unmarshalItem(&ti)
		case `"EndItem"`:
			var ti EndItem
			unmarshalItem(&ti)
		case `"PlatformItem"`:
			var ti PlatformItem
			unmarshalItem(&ti)
		case `"TextItem"`:
			var ti TextItem
			unmarshalItem(&ti)
		case `"PointsItem"`:
			var ti PointsItem
			unmarshalItem(&ti)
		case `"SignalItem"`:
			var ti SignalItem
			unmarshalItem(&ti)
		case `"Place"`:
			var pl Place
			unmarshalItem(&pl)
			delete(sim.TrackItems, pl.ID())
			sim.Places[pl.PlaceCode] = &pl
		default:
			return fmt.Errorf("unknown TrackItem type: %s", rawItem["__type__"])
		}

	}
	if err := sim.checkTrackItemsLinks(); err != nil {
		return err
	}

	sim.Options = rawSim.Options
	sim.Routes = make(map[string]*Route)
	for num, route := range rawSim.Routes {
		route.setSimulation(sim)
		sim.Routes[num] = route
		if err := route.initialize(num); err != nil {
			logger.Crit("error initializing route", "route", route.routeID, "error", err)
			panic(err)
		}
	}
	sim.TrainTypes = rawSim.TrainTypes
	for ttCode, tt := range sim.TrainTypes {
		tt.setSimulation(sim)
		tt.initialize(ttCode)
	}
	sim.Services = rawSim.Services
	for sCode, s := range sim.Services {
		s.setSimulation(sim)
		s.initialize(sCode)
	}
	sim.Trains = rawSim.Trains
	sort.Slice(sim.Trains, func(i, j int) bool {
		switch {
		case len(sim.Trains[i].Service().Lines) == 0 && len(sim.Trains[j].Service().Lines) == 0:
			return sim.Trains[i].ServiceCode < sim.Trains[j].ServiceCode
		case len(sim.Trains[i].Service().Lines) == 0:
			return false
		case len(sim.Trains[j].Service().Lines) == 0:
			return true
		default:
			return sim.Trains[i].Service().Lines[0].ScheduledDepartureTime.Sub(
				sim.Trains[j].Service().Lines[0].ScheduledDepartureTime) < 0
		}
	})
	for i, t := range sim.Trains {
		t.setSimulation(sim, fmt.Sprintf("%d", i))
	}
	sim.MessageLogger = rawSim.MessageLogger
	sim.MessageLogger.setSimulation(sim)
	return nil
}

// MarshalJSON for the Simulation type
func (sim *Simulation) MarshalJSON() ([]byte, error) {
	var res bytes.Buffer
	res.WriteString(`{
	"__type__": "Simulation",
`)
	res.WriteString(`	"messageLogger": `)
	logr, _ := json.Marshal(sim.MessageLogger)
	res.Write(logr)
	res.WriteString(`,
	"Options": `)
	opts, _ := json.Marshal(sim.Options)
	res.Write(opts)
	res.WriteString(`,
	"routes": `)
	rtes, _ := json.Marshal(sim.Routes)
	res.Write(rtes)
	res.WriteString(`,
	"trainTypes": `)
	tts, _ := json.Marshal(sim.TrainTypes)
	res.Write(tts)
	res.WriteString(`,
    "services": `)
	svs, _ := json.Marshal(sim.Services)
	res.Write(svs)

	tkis := make(map[string]TrackItem)
	for k, v := range sim.TrackItems {
		tkis[k] = v
	}
	for _, v := range sim.Places {
		tkis[v.ID()] = v
	}
	res.WriteString(`,
	"trackItems": `)
	tkd, _ := json.Marshal(tkis)
	res.Write(tkd)
	res.WriteString(`,
	"trains": `)
	trns, _ := json.Marshal(sim.Trains)
	res.Write(trns)
	res.WriteString(`,
	"signalLibrary": `)
	sll, _ := json.Marshal(sim.SignalLib)
	res.Write(sll)
	res.WriteString(`}`)
	return res.Bytes(), nil
}

// Initialize initializes the simulation.
// This method must be called before Start.
func (sim *Simulation) Initialize() error {
	sim.MessageLogger.addMessage("Simulation initializing", softwareMsg)
	sim.EventChan = make(chan *Event)
	sim.stopChan = make(chan bool)
	return nil
}

// Start runs the main loop of the simulation by making the clock tick and process each object.
func (sim *Simulation) Start() {
	if sim.stopChan == nil || sim.EventChan == nil {
		panic("You must call Initialize before starting the simulation")
	}
	if sim.started {
		logger.Debug("Simulation already started")
		return
	}
	sim.started = true
	go sim.run()
	logger.Info("Simulation started")
}

// run enters the main loop of the simulation
func (sim *Simulation) run() {
	clockTicker := time.NewTicker(timeStep)
	for {
		select {
		case <-sim.stopChan:
			clockTicker.Stop()
			logger.Info("Simulation paused")
			return
		case <-clockTicker.C:
			sim.increaseTime(timeStep)
			sim.sendEvent(&Event{ClockEvent, sim.Options.CurrentTime})
			sim.updateTrains()
		}
	}
}

// Pause holds the simulation by stopping the clock ticker. Call Start again to restart the simulation.
func (sim *Simulation) Pause() {
	sim.stopChan <- true
	sim.started = false
}

// sendEvent sends the given event on the event channel to notify clients.
// Sending is done asynchronously so as not to block.
func (sim *Simulation) sendEvent(evt *Event) {
	go func() { sim.EventChan <- evt }()
}

// increaseTime adds the step to the simulation time.
func (sim *Simulation) increaseTime(step time.Duration) {
	sim.Options.CurrentTime.Lock()
	defer sim.Options.CurrentTime.Unlock()
	sim.Options.CurrentTime = sim.Options.CurrentTime.Add(time.Duration(sim.Options.TimeFactor) * step)
}

// checks that all TrackItems are linked together.
// Returns the first error met.
func (sim *Simulation) checkTrackItemsLinks() error {
	for _, ti := range sim.TrackItems {
		switch ti.Type() {
		case TypePlace, TypePlatform, TypeText:
			continue
		case TypePoints:
			pi := ti.(*PointsItem)
			if pi.ReverseItem() == nil {
				return ItemNotLinkedAtError{item: ti, pt: pi.Reverse()}
			}
			if !pi.ReverseItem().IsConnected(pi) {
				return ItemInconsistentLinkError{item1: pi, item2: pi.ReverseItem(), pt: pi.Reverse()}
			}
			fallthrough
		case TypeLine, TypeInvisibleLink, TypeSignal:
			if ti.NextItem() == nil {
				return ItemNotLinkedAtError{item: ti, pt: ti.End()}
			}
			if !ti.NextItem().IsConnected(ti) {
				return ItemInconsistentLinkError{item1: ti, item2: ti.NextItem(), pt: ti.End()}
			}
			fallthrough
		case TypeEnd:
			if ti.PreviousItem() == nil {
				return ItemNotLinkedAtError{item: ti, pt: ti.Origin()}
			}
			if !ti.PreviousItem().IsConnected(ti) {
				return ItemInconsistentLinkError{item1: ti, item2: ti.PreviousItem(), pt: ti.End()}
			}
		}
	}
	return nil
}

// updateTrains update all trains information such as status, position, speed, etc.
func (sim *Simulation) updateTrains() {
	for _, train := range sim.Trains {
		if !train.IsActive() {
			continue
		}
		train.activate(sim.Options.CurrentTime)
		train.advance(timeStep)
	}
}

// RegisterRoutesManager registers the given route manager in the simulation.
//
// When several routes managers are registered, all of them are called in turn.
// If all of them respond true, then the response is true. If one responds false,
// the response is false.
func RegisterRoutesManager(rm RoutesManager) {
	routesManagers = append(routesManagers, rm)
}

// RegisterTrainsManager registers the given trains manager in the simulation.
//
// There can be several trains managers registered, but each train will use only one.
// If a train has not been explicitly set to a trains manager, it will use the default
// one. Default trains manager is the first registered manager.
func RegisterTrainsManager(tm TrainsManager) {
	if trainsManagers == nil {
		trainsManagers = make(map[string]TrainsManager)
		defaultTrainManager = tm
	}
	trainsManagers[tm.Name()] = tm
}

// RegisterLineItemManager registers the given line manager in the simulation.
//
// If a line manager was already registered, it is replaced by lim.
func RegisterLineItemManager(lim LineItemManager) {
	lineItemManager = lim
}

// RegisterPointsItemManager registers the given points manager in the simulation.
//
// If a points manager was already registered, it is replaced by pim.
func RegisterPointsItemManager(pim PointsItemManager) {
	pointsItemManager = pim
}

// RegisterSignalItemManager registers the signal manager in the simulation.
//
// If a signals manager was already registered, it is replaced by sim.
func RegisterSignalItemManager(sim SignalItemManager) {
	signalItemManager = sim
}
