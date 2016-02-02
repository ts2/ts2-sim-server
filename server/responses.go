/*   Copyright (C) 2008-2016 by Nicolas Piganeau and the TS2 team
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

package server

import (
	"fmt"
	"github.com/ts2/ts2-sim-server/simulation"
)

type StatusCode string

const (
	OK StatusCode = "OK"
	FAIL StatusCode = "FAIL"
)

type MessageType string

const (
	RESPONSE MessageType = "response"
	EVENT    MessageType = "event"
	REQUEST  MessageType = "request"
)

/*
DataStatus is the Data part of a ResponseStatus message
*/
type DataStatus struct {
	Status  StatusCode `json:"status"`
	Message string     `json:"message"`
}

/*
ResponseStatus is a status message sent to a websocket client
*/
type ResponseStatus struct {
	MsgType MessageType `json:"msgType"`
	Data    DataStatus  `json:"data"`
}

/*
DataEvent is the Data part of a ResponseEvent message
*/
type DataEvent struct {
	Name   simulation.EventName `json:"name"`
	Object interface{}          `json:"object"`
}

/*
ResponseEvent is a message sent by the server to the clients when an event is triggered in the simulation
*/
type ResponseEvent struct {
	MsgType MessageType `json:"msgType"`
	Data    DataEvent   `json:"data"`
}

/*
NewErrorResponse returns a ResponseStatus object corresponding to the given error.
*/
func NewErrorResponse(e error) *ResponseStatus {
	sr := ResponseStatus{
		MsgType: RESPONSE,
		Data: DataStatus{
			FAIL,
			fmt.Sprintf("Error: %s", e),
		},
	}
	return &sr
}

/*
NewOkResponse returns a new ResponseStatus object with OK status and empty message.
*/
func NewOkResponse(msg string) *ResponseStatus {
	sr := ResponseStatus{
		MsgType: RESPONSE,
		Data: DataStatus{
			OK,
			msg,
		},
	}
	return &sr
}

/*
NewEventResponse returns a new ResponseEvent object from the given Event
 */
func NewEventResponse(e *simulation.Event) *ResponseEvent {
	er := ResponseEvent{
		MsgType: EVENT,
		Data: DataEvent{
			Name: e.Name,
			Object: e.Object,
		},
	}
	return &er
}