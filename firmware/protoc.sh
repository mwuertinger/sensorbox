#!/bin/bash
set -e
nanopb_generator.py -Dsrc -I.. ../config.proto
