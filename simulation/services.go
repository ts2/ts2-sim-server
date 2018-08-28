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

type serviceActionCode uint8

const (
	// actionReverse the train
	actionReverse serviceActionCode = 1

	// actionSetService set the given service. ActionParam is the new service code
	actionSetService serviceActionCode = 2

	// actionSplit the train at the given position. ActionParam is the element after
	// which to split (integer).
	actionSplit serviceActionCode = 3

	// actionJoin the train. ActionParam is 'ahead' if to join with the train in
	// front or 'behind' otherwise.
	actionJoin serviceActionCode = 4
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
	TrackCode              string `json:"TrackCode"`

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
	Description          string           `json:"description"`
	Lines                []*ServiceLine   `json:"lines"`
	PlannedTrainTypeCode string           `json:"plannedTrainType"`
	PostActions          []*ServiceAction `json:"postActions"`

	simulation *Simulation
}

// PlannedTrainType returns a pointer to the planned TrainType for this Service.
func (s *Service) PlannedTrainType() *TrainType {
	// TODO catch error
	return s.simulation.TrainTypes[s.PlannedTrainTypeCode]
}

// setSimulation() sets a pointer to the Simulation this Service to be part of
func (s *Service) setSimulation(sim *Simulation) {
	s.simulation = sim
	for _, line := range s.Lines {
		line.service = s
	}
}
