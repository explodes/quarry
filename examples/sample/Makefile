

dirs:
	mkdir -p ./samplepb || true

proto: dirs data.proto
	protoc --go_out=./samplepb data.proto

sample: proto
	go build -o sample ./cmd/sample/sample.go

clean:
	rm -rf ./samplepb || true
	rm sample || true