#!/bin/bash
cd ..
for pattern in ../patterns/*.noise; do
	node noiseExplorer \
		--generate=go --pattern=$pattern \
		> ../implementations/$(basename "${pattern}").go
done
cd util