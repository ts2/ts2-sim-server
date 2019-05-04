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
	"fmt"

	"github.com/ts2/ts2-sim-server/simulation"
)

type StatusCode string

const (
	Ok   StatusCode = "OK"
	Fail StatusCode = "FAIL"
)

// A MessageType defines the type of a JSON message on websocket
type MessageType string

const (
	TypeResponse     MessageType = "response"
	TypeNotification MessageType = "notification"
)

// Response is a status message sent to a websocket client
type Response struct {
	ID      int         `json:"id"`
	MsgType MessageType `json:"msgType"`
	Data    RawJSON     `json:"data"`
}

// DataStatus is the Data part of a ResponseStatus message
type DataStatus struct {
	Status  StatusCode `json:"status"`
	Message string     `json:"message"`
}

// ResponseStatus is a status message sent to a websocket client
type ResponseStatus struct {
	ID      int         `json:"id"`
	MsgType MessageType `json:"msgType"`
	Object  string      `json:"object"`
	Action  string      `json:"action"`
	Data    DataStatus  `json:"data"`
}

// DataEvent is the Data part of a ResponseNotification message
type DataEvent struct {
	Name   simulation.EventName `json:"name"`
	Object interface{}          `json:"object"`
}

// ResponseNotification is a message sent by the server to the clients when an event is triggered in the simulation
type ResponseNotification struct {
	MsgType MessageType `json:"msgType"`
	Data    DataEvent   `json:"data"`
}

// NewResponse returns a Response with the given data
func NewResponse(id int, data RawJSON) *Response {
	r := Response{
		ID:      id,
		MsgType: TypeResponse,
		Data:    data,
	}
	return &r
}

// NewErrorResponse returns a ResponseStatus object corresponding to the given error.
func NewErrorResponse(id int, e error) *ResponseStatus {
	sr := ResponseStatus{
		ID:      id,
		MsgType: TypeResponse,
		Data: DataStatus{
			Fail,
			fmt.Sprintf("Error: %s", e),
		},
	}
	return &sr
}

// NewOkResponse returns a new ResponseStatus object with OK status and empty message.
func NewOkResponse(id int, obj string, action string, msg string) *ResponseStatus {
	sr := ResponseStatus{
		ID:      id,
		MsgType: TypeResponse,
		Object:  obj,
		Action:  action,
		Data: DataStatus{
			Ok,
			msg,
		},
	}
	return &sr
}

// NewNotificationResponse returns a new ResponseNotification object from the given Event
func NewNotificationResponse(e *simulation.Event) *ResponseNotification {
	er := ResponseNotification{
		MsgType: TypeNotification,
		Data: DataEvent{
			Name:   e.Name,
			Object: e.Object,
		},
	}
	return &er
}
