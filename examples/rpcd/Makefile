

dirs:
	mkdir -p ./rpcdpb || true

proto: dirs rpcd.proto
	protoc --go_out=plugins=grpc:./rpcdpb rpcd.proto

userd: proto
	go build -o rpcd_userd ./cmd/userd/userd.go

userc: proto
	go build -o rpcd_userc ./cmd/userc/userc.go

clean:
	rm -rf ./rpcdpb || true
	rm rpcd_userd || true
	rm rpcd_userc || true