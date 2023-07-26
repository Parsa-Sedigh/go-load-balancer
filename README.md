### How to run the project?
```shell
make run_demo_servers

# on another terminal window
make run_round_robin # or run `make run_weighted_round_robin`
```

### Notes:
- Unit tests are co-located with the related files
- Integration tests are at `cmd/demo/server_test.go`
- to run the `TestRoundRobin` test, run the load balancer in round-robin mode using `make run_round_robin` and to run the `TestWeightedRoundRobin` test,
run the load balancer in weighted-round-robin mode using `make run_weighted_round_robin`