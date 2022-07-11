#!/bin/bash
set -ex
pwd
nanopb_generator.py -D src -I .. ../sensorbox.proto
