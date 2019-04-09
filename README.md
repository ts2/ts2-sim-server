ts2 Sim Server
==============

[![GoDoc](https://godoc.org/github.com/ts2/ts2-sim-server?status.svg)](https://godoc.org/github.com/ts2/ts2-sim-server)
[![Join the chat at https://gitter.im/ts2/ts2-sim-server](https://badges.gitter.im/ts2/ts2-sim-server.svg)](https://gitter.im/ts2/ts2-sim-server?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)
[![Build Status](https://secure.travis-ci.org/ts2/ts2-sim-server.svg)](http://travis-ci.org/ts2/ts2-sim-server)
[![codecov](https://codecov.io/gh/ts2/ts2-sim-server/branch/master/graph/badge.svg)](https://codecov.io/gh/ts2/ts2-sim-server)

This is the home of the ts2 simulation server, the core of the ts2 simulator.

Unless you want to develop your own client and access the simulator through its API, 
you should go to https://github.com/ts2/ts2 to grab the all-in-one simulator which includes the simulation server.


Install
-------
```bash
go get github.com/ts2/ts2-sim-server
```

Starting the server
-------------------
```bash
ts2-sim-server /path/to/simulation-file.json
```
The server is running and can be accessed at `ws://localhost:22222/ws`

Web UI
------
The server ships with a minimal Web UI to interact with the webservice.

Start the server and head to `http://localhost:22222`.
