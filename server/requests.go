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

package server

import (
	"encoding/json"
	"errors"

	"github.com/ts2/ts2-sim-server/simulation"
)

// RawJSON is a json.RawMessage (i.e. []byte) that prints itself
// as a string to be human readable in logs
type RawJSON json.RawMessage

// String function for the RawJSON type
func (r RawJSON) String() string {
	return string(r)
}

// MarshalJSON returns m as the JSON encoding of r.
func (r RawJSON) MarshalJSON() ([]byte, error) {
	if r == nil {
		return []byte("null"), nil
	}
	return r, nil
}

// UnmarshalJSON sets *r to a copy of data.
func (r *RawJSON) UnmarshalJSON(data []byte) error {
	if r == nil {
		return errors.New("RawJSON: UnmarshalJSON on nil pointer")
	}
	*r = append((*r)[0:0], data...)
	return nil
}

// Request is a generic request made by a websocket client.
//
// It is used before dispatching and unmarshaling into a specific request type.
type Request struct {
	ID     int     `json:"id"`
	Object string  `json:"object"`
	Action string  `json:"action"`
	Params RawJSON `json:"params"`
}

// ParamsRegister is the struct of the Request Params for a RequestRegister
type ParamsRegister struct {
	ClientType    ClientType  `json:"type"`
	ClientSubType ManagerType `json:"subType"`
	Token         string      `json:"token"`
}

// RequestRegister is a request made by a websocket client to log onto the server.
type RequestRegister struct {
	ID     int            `json:"id"`
	Object string         `json:"object"`
	Action string         `json:"action"`
	Params ParamsRegister `json:"params"`
}

// ParamsListener is the struct of the Request Params for a RequestListener
type ParamsListener struct {
	Event simulation.EventName `json:"event"`
	IDs   []string             `json:"ids"`
}

// RequestListener is a request made by a websocket client to add or remove a listener.
type RequestListener struct {
	ID     int            `json:"id"`
	Object string         `json:"object"`
	Action string         `json:"action"`
	Params ParamsListener `json:"params"`
}
