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

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"

	_ "github.com/ts2/ts2-sim-server/plugins/points"
	_ "github.com/ts2/ts2-sim-server/plugins/routes"
	_ "github.com/ts2/ts2-sim-server/plugins/signals"
	_ "github.com/ts2/ts2-sim-server/plugins/trains"
	"github.com/ts2/ts2-sim-server/server"
	"github.com/ts2/ts2-sim-server/simulation"
	log "gopkg.in/inconshreveable/log15.v2"
)

var logger log.Logger

func main() {
	// Command line arguments
	port := flag.String("port", server.DefaultPort, "The port on which the server will listen")
	addr := flag.String("addr", server.DefaultAddr, "The address on which the server will listen. Set to 0.0.0.0 to listen on all addresses.")
	logFile := flag.String("logfile", "", "The filename in which to save the logs. If not specified, the logs are sent to stderr.")
	logLevel := flag.String("loglevel", "info", "The minimum level of log to be written. Possible values are 'crit', 'error', 'warn', 'info' and 'debug'.")

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

	// Setup logging system
	logger = log.New()
	var outputHandler log.Handler
	if *logFile != "" {
		outputHandler = log.Must.FileHandler(*logFile, log.LogfmtFormat())
	} else {
		outputHandler = log.StderrHandler
	}
	logLvl, err_level := log.LvlFromString(*logLevel)
	if err_level != nil {
		fmt.Fprintf(os.Stderr, "Error: Unknown loglevel\n\n")
		flag.Usage()
		os.Exit(1)
	}
	logger.SetHandler(log.LvlFilterHandler(
		logLvl,
		outputHandler,
	))
	simulation.InitializeLogger(logger)
	server.InitializeLogger(logger)

	// Load the simulation
	if len(flag.Args()) == 0 {
		fmt.Fprintf(os.Stderr, "Error: Please specify a simulation file\n\n")
		flag.Usage()
		os.Exit(1)
	}
	simFile := flag.Arg(0)
	logger.Info("Loading simulation", "file", simFile)

	data, err := ioutil.ReadFile(simFile)
	if err != nil {
		logger.Crit("Unable to read file", "file", simFile, "error", err)
		os.Exit(1)
	}

	var sim simulation.Simulation
	if err = json.Unmarshal(data, &sim); err != nil {
		logger.Error("Load Error", "file", simFile, "error", err)
		return
	}

	if err = sim.Initialize(); err != nil {
		logger.Error("Invalid simulation", "file", simFile, "error", err)
		return
	}
	logger.Info("Simulation loaded", "sim", sim.Options.Title)

	go server.Run(&sim, *addr, *port)

	select {
	case <-killChan:
		// TODO gracefully shutdown things maybe
		logger.Info("Server killed, exiting...")
		os.Exit(0)
	}
}
