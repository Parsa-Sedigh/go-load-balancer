.PHONY: run_demo_servers run_round_robin run_weighted_round_robin

run_round_robin:
	go run cmd/load_balancer/main.go --config-path=example/config.yaml

run_weighted_round_robin:
	go run cmd/load_balancer/main.go --config-path=example/config-weighted.yaml

run_demo_servers:
	go run cmd/demo/server.go & go run cmd/demo/server.go --port=8082