#!/bin/bash
set -e
nanopb_generator.py -D src -I .. ../config.proto
