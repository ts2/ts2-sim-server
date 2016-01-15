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
	"html/template"
	"net/http"
)


// getTemplate() relies on go-bindata being run.. in -debug is dynamic
func getTemplate(name string) *template.Template {
	return template.Must(template.New(name).Parse(string(MustAsset(name))))
}

var tplHome *template.Template
var TplWhateverFooBar *template.Template

func init(){
	//tplHome = getTemplate("templates/home.html")
}


/*
H_Home()  handles and serves home.html page with integrated JS WebSocket client.
*/
func H_HomePage(w http.ResponseWriter, r *http.Request) {
	//if r.URL.Path != "/" {
	//	http.Error(w, "404: Not found", 404)
	//	return
	//}
	if r.Method != "GET" {
		http.Error(w, "405: Method not allowed", 405)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	data := struct {
		Title       string
		Description string
		Host        string
	}{
		sim.Options.Title,
		sim.Options.Description,
		"ws://" + r.Host + "/ws",
	}

	// For now we compile every time = but neeed "Dev" flag here..
	tplHome = getTemplate("templates/home.html")
	tplHome.Execute(w, data)
}



