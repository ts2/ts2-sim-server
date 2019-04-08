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

type serviceActionCode string

const (
	// actionReverse the train
	actionReverse serviceActionCode = "REVERSE"

	// actionSetService set the given service. ActionParam is the new service code
	actionSetService serviceActionCode = "SET_SERVICE"

	// actionSplit the train at the given position. ActionParam is the element after
	// which to split (integer).
	actionSplit serviceActionCode = "SPLIT"

	// actionJoin the train. ActionParam is 'ahead' if to join with the train in
	// front or 'behind' otherwise.
	actionJoin serviceActionCode = "JOIN"
)

// A ServiceAction is an action that can be performed on a train
type ServiceAction struct {
	ActionCode  serviceActionCode `json:"actionCode"`
	ActionParam string            `json:"actionParam"`
}

// ServiceLine is a line of the definition of the Service.
//
// It consists of a TypePlace (usually a station) with a track number
// and scheduled times to arrive at and depart from this station.
type ServiceLine struct {
	MustStop               bool   `json:"mustStop"`
	PlaceCode              string `json:"placeCode"`
	ScheduledArrivalTime   Time   `json:"scheduledArrivalTime"`
	ScheduledDepartureTime Time   `json:"scheduledDepartureTime"`
	TrackCode              string `json:"trackCode"`

	service *Service
}

// Place associated with this service line
func (sl *ServiceLine) Place() *Place {
	return sl.service.simulation.Places[sl.PlaceCode]
}

// A Service is mainly a predefined schedule that trains are supposed to
// follow with a few additional informations.
//
// The schedule is composed of several "lines" of type ServiceLine
type Service struct {
	serviceID            string
	Description          string           `json:"description"`
	Lines                []*ServiceLine   `json:"lines"`
	PlannedTrainTypeCode string           `json:"plannedTrainType"`
	PostActions          []*ServiceAction `json:"postActions"`

	simulation *Simulation
}

// ID returns the unique identifier of this service: its code
func (s *Service) ID() string {
	return s.serviceID
}

// PlannedTrainType returns a pointer to the planned TrainType for this Service.
func (s *Service) PlannedTrainType() *TrainType {
	// TODO catch error
	return s.simulation.TrainTypes[s.PlannedTrainTypeCode]
}

// setSimulation sets a pointer to the Simulation this Service to be part of
func (s *Service) setSimulation(sim *Simulation) {
	s.simulation = sim
}

// initialize the current service
func (s *Service) initialize(code string) {
	s.serviceID = code
	for _, line := range s.Lines {
		line.service = s
	}
}

// MarshalJSON for the Service type
func (s *Service) MarshalJSON() ([]byte, error) {
	type auxService struct {
		ID                   string           `json:"id"`
		Description          string           `json:"description"`
		Lines                []*ServiceLine   `json:"lines"`
		PlannedTrainTypeCode string           `json:"plannedTrainType"`
		PostActions          []*ServiceAction `json:"postActions"`
	}
	as := auxService{
		ID:                   s.ID(),
		Description:          s.Description,
		Lines:                s.Lines,
		PlannedTrainTypeCode: s.PlannedTrainTypeCode,
		PostActions:          s.PostActions,
	}
	d, err := json.Marshal(as)
	return d, err
}
