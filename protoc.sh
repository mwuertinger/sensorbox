set -e
mkdir -p server/pb
mkdir -p config-tool/pb
protoc -I. -I/usr/lib/python3/dist-packages/proto --go_out=paths=source_relative:server/pb --go_opt=Mnanopb.proto=github.com/mwuertinger/sensorbox/server/pb sensorbox.proto
protoc -I. -I/usr/lib/python3/dist-packages/proto --go_out=paths=source_relative:config-tool/pb --go_opt=Mnanopb.proto=github.com/mwuertinger/sensorbox/config-tool/pb sensorbox.proto
