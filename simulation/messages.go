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

// MessageType defines the type of message of the Logger
type MessageType uint8

const (
	softwareMsg      MessageType = 0
	playerWarningMsg MessageType = 1
	simulationMsg    MessageType = 2
)

// Message is one message emitted to the MessageLogger of the simulation.
type Message struct {
	MsgType MessageType `json:"msgType"`
	MsgText string      `json:"msgText"`
}

// MessageLogger holds all Message instances that have been emitted to it.
type MessageLogger struct {
	Messages   []Message `json:"messages"`
	simulation *Simulation
}

// setSimulation() sets the Simulation this MessageLogger is part of.
func (ml *MessageLogger) setSimulation(sim *Simulation) {
	ml.simulation = sim
}

// addMessage adds the given message to the simulation message Logger.
// This method also logs to the Logger the same message.
func (ml *MessageLogger) addMessage(msg string, typ MessageType) {
	ml.Messages = append(ml.Messages, Message{
		MsgText: msg,
		MsgType: typ,
	})
	if Logger != nil {
		Logger.Info(msg, "msgType", typ)
	}
}
