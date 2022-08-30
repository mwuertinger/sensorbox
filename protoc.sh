#!/bin/bash
set -e

for dir in server/pb config-tool/pb
do
	mkdir -p $dir
	# nanopb is only relevant for the firmware and irrelevant in Go. Therefore a fake package is set here and then removed with sed below.
	protoc -I. -I/usr/lib/python3/dist-packages/proto --go_out=paths=source_relative:$dir --go_opt=Mnanopb.proto=to/remove sensorbox.proto
	sed -i '/_ "to\/remove"/d' $dir/sensorbox.pb.go
done

nanopb_generator.py -D firmware/src -I . sensorbox.proto
