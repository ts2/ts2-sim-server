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
	"encoding/json"
	"net/http"
)

func NewAjaxPayload() map[string]interface{} {
	pay := make(map[string]interface{})
	pay["error"] = ""
	return pay
}

/*
H_Ajax() - handles and serves ajax requests for "/ajax".
*/
func H_AjaxIndex(w http.ResponseWriter, r *http.Request) {
	data := struct {
		Title       string
		Description string
		Host        string
		Urls		[]string
	}{
		sim.Options.Title,
		sim.Options.Description,
		"ws://" + r.Host + "/ws",
		[]string{"/trains", "/foo"},
	}

	SendJson(w, data)
}

/*
H_AjaxTrains() - handles and serves ajax requests for "/ajax/trains" - test.
*/
func H_AjaxTrains(w http.ResponseWriter, r *http.Request) {

	data := NewAjaxPayload()
	data["trains"] = sim.Trains
	SendJson(w, data)
}

func SendJson(w http.ResponseWriter, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	var err error
	var json_data []byte

	json_data, err = json.MarshalIndent(payload, "", "  ")
	if err != nil {
		json_data, err = json.MarshalIndent(NewErrorResponse(err), "", "")
	}
	w.Write(json_data)
}
