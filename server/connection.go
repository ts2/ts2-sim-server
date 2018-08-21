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
	"context"
	"encoding/json"
	"fmt"

	"github.com/gorilla/websocket"
)

type ClientType string

const (
	Client ClientType = "client"
)

type ManagerType string

// connection is a wrapper around the websocket.Conn
type connection struct {
	websocket.Conn
	// pushChan is the channel on which pushed messaged are sent
	pushChan    chan interface{}
	clientType  ClientType
	ManagerType ManagerType
	LastRequest Request
}

// loop starts the reading and writing loops of the connection.
func (conn *connection) loop(ctx context.Context) {
	logger.Debug("New connection", "remote", conn.RemoteAddr())
	if err, req := conn.registerClient(); err != nil {
		// Try to notify client
		conn.WriteJSON(NewErrorResponse(req.ID, err))
		logger.Error("Error while login", "connection", conn.RemoteAddr(), "error", err)
		return
	}
	loopCtx, childCancel := context.WithCancel(ctx)
	defer childCancel()
	go conn.processWrite(loopCtx)
	conn.processRead(loopCtx)
}

// processRead performs all read operations from the connection and forwards to the hub
func (conn *connection) processRead(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		err := conn.ReadJSON(&conn.LastRequest)
		if err != nil {
			if _, ok := err.(*websocket.CloseError); ok {
				logger.Debug("Connection closed by peer", "connection", conn.RemoteAddr())
				conn.Close()
				return
			} else {
				logger.Info("Error while reading", "connection", conn.RemoteAddr(), "error", err)
				conn.pushChan <- NewErrorResponse(conn.LastRequest.ID, err)
				continue
			}
		}
		hub.readChan <- conn
	}
}

// processWrite performs all the write operations to the connection sent by the hub
func (conn *connection) processWrite(ctx context.Context) {
	for {
		select {
		case req := <-conn.pushChan:
			if err := conn.WriteJSON(req); err != nil {
				logger.Info("Error while writing", "connection", conn.RemoteAddr(), "request", req, "error", err)
			}
		case <-ctx.Done():
			return
		}
	}
}

// registerClient() waits for a register request from the client, checks it and registers the connection
// on the hub if it is valid. Otherwise it returns an error.
func (conn *connection) registerClient() (error, *Request) {
	// Parse request
	req := new(Request)
	if err := conn.ReadJSON(req); err != nil {
		return err, req
	}
	if req.Object != "server" || req.Action != "register" {
		return fmt.Errorf("register required"), req
	}
	registerParams := ParamsRegister{}
	if err := json.Unmarshal(req.Params, &registerParams); err != nil {
		return fmt.Errorf("unable to parse register params: %s", err), req
	}

	// Authenticate client and type
	if registerParams.ClientType == Client &&
		registerParams.Token == sim.Options.ClientToken {
		conn.clientType = Client
	} else {
		return fmt.Errorf("invalid register parameters"), req
	}

	// authenticated, so setup
	if err := conn.WriteJSON(NewOkResponse(req.ID, "Successfully registered")); err != nil {
		logger.Info("Error while writing", "connection", conn.RemoteAddr(), "request", "NewOkResponse", "error", err)
	}
	hub.registerChan <- conn
	logger.Info("Registered", "connection", conn.RemoteAddr(), "clientType", conn.clientType, "managerType", conn.ManagerType)
	return nil, req
}

// Close terminates the websocket connection and closes associated resources
func (conn *connection) Close() error {
	conn.Conn.Close()
	hub.unregisterChan <- conn
	return nil
}
