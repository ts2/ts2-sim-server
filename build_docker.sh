#!/usr/bin/env bash
cp ts2-sim-server docker/
cp simulation/testdata/demo.json docker/
docker build -t ts2simulator/sim-server --no-cache docker
