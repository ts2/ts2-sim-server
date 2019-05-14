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
	"io/ioutil"
	"os"
	"testing"

	log "gopkg.in/inconshreveable/log15.v2"
)

func TestMain(m *testing.M) {
	mainLogger := log.New()
	if os.Getenv("TS2_DEBUG") == "" {
		mainLogger.SetHandler(log.DiscardHandler())
	}
	InitializeLogger(mainLogger)
	InitializeLogger(mainLogger)
	os.Exit(m.Run())
}

// Equals is a comparison function for DelayGenerator objects.
func (dg DelayGenerator) Equals(b DelayGenerator) bool {
	for i := 0; i < len(dg.data); i++ {
		if dg.data[i] != b.data[i] {
			return false
		}
	}
	return true
}

func loadSim(filename string) []byte {
	data, _ := ioutil.ReadFile(filename)
	return data
}
