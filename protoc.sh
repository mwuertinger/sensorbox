#!/bin/bash
# Generates protobuf code for the server, the config-tool and the firmware.
#
# Usage: ./protoc.sh [server] [config-tool] [firmware]   (default: all of them)
#
# Requirements:
#   pip install nanopb==0.4.9.1 grpcio-tools  (nanopb must match the firmware runtime, see firmware/platformio.ini)
#   Go for the server and config-tool targets. protoc-gen-go is built from the google.golang.org/protobuf
#   version in the module's go.mod, so the generated code always matches the runtime it is compiled against.
set -euo pipefail
cd "$(dirname "$0")"

targets=("$@")
[ ${#targets[@]} -gt 0 ] || targets=(server config-tool firmware)

NANOPB_PROTO_DIR="$(python3 -c 'import nanopb, os; print(os.path.join(os.path.dirname(nanopb.__file__), "generator", "proto"))')"
GRPC_PROTO_DIR="$(python3 -c 'import grpc_tools, os; print(os.path.join(os.path.dirname(grpc_tools.__file__), "_proto"))')"

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT

for target in "${targets[@]}"
do
	case "$target" in
	server|config-tool)
		(cd "$target" && go build -o "$tmp/$target/protoc-gen-go" google.golang.org/protobuf/cmd/protoc-gen-go)
		mkdir -p "$target/pb"
		# nanopb is only relevant for the firmware and irrelevant in Go. Therefore a fake package is set here and then removed with sed below.
		python3 -m grpc_tools.protoc -I. -I"$NANOPB_PROTO_DIR" -I"$GRPC_PROTO_DIR" \
			--plugin=protoc-gen-go="$tmp/$target/protoc-gen-go" \
			--go_out=paths=source_relative:"$target/pb" --go_opt=Mnanopb.proto=to/remove sensorbox.proto
		sed -i '/_ "to\/remove"/d' "$target/pb/sensorbox.pb.go"
		;;
	firmware)
		nanopb_generator -D firmware/src -I . sensorbox.proto
		;;
	*)
		echo "unknown target: $target (expected server, config-tool or firmware)" >&2
		exit 2
		;;
	esac
done
