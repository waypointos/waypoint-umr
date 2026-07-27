test:
	go test ./...

# Needs a dev rover in another terminal: cd ../waypoint && make dev-rover
dev:
	WAYPOINT_NATS_URL=nats://127.0.0.1:4222 \
	WAYPOINT_ROVER_ID=sim-rover \
	WAYPOINT_MODULE_ID=umr \
	go run ./cmd/waypoint-module-umr

raw:
	$(MAKE) -f build/Makefile raw
