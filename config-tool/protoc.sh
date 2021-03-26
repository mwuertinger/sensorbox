set -e
mkdir -p pb
protoc -I../firmware/src -I../nanopb/generator/proto --go_out=paths=source_relative:pb ../firmware/src/config.proto
