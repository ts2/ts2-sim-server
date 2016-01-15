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

// Main ts2-tim-server command
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/ts2/ts2-sim-server/server"
	"github.com/ts2/ts2-sim-server/simulation"
)

func main() {
	// Command line arguments
	debug := flag.Bool("debug", false, "Enable debugging")
	port := flag.String("port", server.DEFAULT_PORT, "The port on which the server will listen")
	addr := flag.String("addr", server.DEFAULT_ADDR, "The address on which the server will listen. Set to 0.0.0.0 to listen on all addresses.")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `Usage of ts2-sim-server:
  ts2-sim-server [options...] file

ARGUMENTS:
  file
		The JSON simulation file to load

OPTIONS:
`)
		flag.PrintDefaults()
	}

	flag.Parse()

	// Handle ctrl+c to kill on terminal
	killChan := make(chan os.Signal, 1)
	signal.Notify(killChan, os.Interrupt)

	// Load the simulation
	if len(flag.Args()) == 0 {
		fmt.Fprintf(os.Stderr, "Error: Please specify a simulation file\n\n")
		flag.Usage()
		os.Exit(1)
	}
	simFile := flag.Arg(0)
	log.Printf("Loading simulation: %s\n", simFile)
	/*
	data, err := ioutil.ReadFile(simFile)
	if err != nil {
		log.Fatal(err)
	}
	*/

	var sim simulation.Simulation
	sim.Debug = debug
	errload := sim.Load(simFile)
	//errload := json.Unmarshal(data, &sim)

	if errload != nil {
		log.Printf("Load Error: %s\n", errload)
		return
	}
	log.Printf("Simulation loaded: %s\n", sim.Options.Title)


	go server.Run(&sim, *addr, *port)

	// Route all messages
	for {
		select {

		case <-killChan:
			// TODO gracefully shutdown things maybe
			log.Println("Server killed, exiting...")
			os.Exit(0)
		}
	}
}
