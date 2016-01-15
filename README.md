ts2 Sim Server
====================================

[![GoDoc](https://godoc.org/github.com/ts2/ts2-sim-server?status.svg)](https://godoc.org/github.com/ts2/ts2-sim-server)
[![Join the chat at https://gitter.im/ts2/ts2-sim-server](https://badges.gitter.im/ts2/ts2-sim-server.svg)](https://gitter.im/ts2/ts2-sim-server?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)
[![Build Status](https://secure.travis-ci.org/ts2/ts2-sim-server.svg?branch=pedro-dev)](http://travis-ci.org/revel/revel) 


This is code for the new TS2 sim server, written in golang

- Please visit https://github.com/ts2/ts2/wiki/New-Arch
- https://github.com/ts2/ts2/wiki/TS2-Sim-Server-Specifications
- Under development as v0.7 "next architecture"



Notes
-------------

```
// install go-bindata 
go get github.com/jteeuwen/go-bindata/...

// create asset files
go-bindata -debug -pkg server -o server/bindata_templates.go templates/
```

0.7 - latest
--------------------
Npi has almost rewitten code in golang [whaw] ;-)
