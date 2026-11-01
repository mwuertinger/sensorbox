#!/bin/bash
# Generates protobuf code for the server, the config-tool and the firmware.
#
# Requirements:
#   pip install nanopb==0.4.9.1 grpcio-tools  (nanopb must match the firmware runtime, see firmware/platformio.ini)
#   go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
set -euo pipefail
cd "$(dirname "$0")"

NANOPB_PROTO_DIR="$(python3 -c 'import nanopb, os; print(os.path.join(os.path.dirname(nanopb.__file__), "generator", "proto"))')"
GRPC_PROTO_DIR="$(python3 -c 'import grpc_tools, os; print(os.path.join(os.path.dirname(grpc_tools.__file__), "_proto"))')"

for dir in server/pb config-tool/pb
do
	mkdir -p "$dir"
	# nanopb is only relevant for the firmware and irrelevant in Go. Therefore a fake package is set here and then removed with sed below.
	python3 -m grpc_tools.protoc -I. -I"$NANOPB_PROTO_DIR" -I"$GRPC_PROTO_DIR" \
		--go_out=paths=source_relative:"$dir" --go_opt=Mnanopb.proto=to/remove sensorbox.proto
	sed -i '/_ "to\/remove"/d' "$dir/sensorbox.pb.go"
done

nanopb_generator -D firmware/src -I . sensorbox.proto
