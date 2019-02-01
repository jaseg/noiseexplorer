#!/bin/bash
cd ..
for pattern in ../patterns/*.noise; do
    node noiseExplorer \
		--generate=go --testgen --pattern=$pattern \
		> ../implementations/go/tests/$(basename "${pattern}").go
done
echo " OK"
echo "[NoiseExplorer] Running Tests..."
cd ../implementations/go/tests
go get -d ./...
cd ../../../src
for pattern in ../implementations/go/tests/*.go; do
    go run $pattern
done
cd util
