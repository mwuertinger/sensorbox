#!/bin/bash
set -ex
nanopb_generator.py -D src -I .. ../config.proto
