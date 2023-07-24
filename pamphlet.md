## Part 1

Farely == load balancer

The Strategy field can be either in the `Config` struct(therefore, all of the services have the same load balancing strategy) or 
in the `Service` struct(so having a load balancing strategy per service). But let's keep it simple for now, so for now we go with the first
option.

The `Next` method on `ServerList`, gives us the next server.

We have two concerns about the current `ServerList` `Next`:
1. the modulo operator(%) is not the fastest operator! In order to not use the modulo operator, we increment the nxt(atomically) and then
if it's more than or equal to the number of services that we have, subtract the next from the current number of services:
```go
if nxt >= lenS {
	nxt -= lenS
}
```
2. it's not concurrent safe. Because we can have 2 servers calling this method at the same time and we'll have a missed increment, so the load
will be doubled on one given server. So we should do the increment in the `Next` method atomically using `atomic.AddInt32`.

We have a bunch of `Server`s per `Service` name. For now, we're ignoring the service name and the required strategy per service name.

To test the demo server directly:
```shell
# Run this to start the demo server
go run cmd/demo/test_server.go

# then run:
curl localhost:8081
```

To test the proxy:
```shell
go run cmd/load_balancer/main.go
go run cmd/demo/test_server.go

curl localhost:8080 # send request to the load balancer server
```

### Add more replicas of a server
To start more **replicas** of the demo server that we have in `cmd/demo/test_server`, just open new terminal windows and run:
```shell
go run cmd/demo/test_server.go --port=<new port>
```

Then register that replica in Config struct in `cmd/load_balancer/main.go`.

**Note:** Now it's gonna load balance between all of the replicas. But the problem is if one of the replicas is down, the
load balancer is gonna still proxy the req to that down replica and it's not gonna detect that that replica is down.

Now we need to load the configuration using yaml.
```shell
go get gopkg.in/yaml.v2
```

Now to test things:
```shell
# run the load balancer
go run cmd/load_balancer/main.go --config-path=example/config.yaml

# run the servers to balance the load to them
go run cmd/demo/test_server.go
go run cmd/demo/test_server.go --port=8082

# send a req to load balancer(running on port 8080)
curl localhost:8080
```

## Part 2
To run the load balancer, from the root of the project run:
```shell
go run cmd/load_balancer/main.go --config-path ./example/config.yaml

# let's run two servers for load balancing
go run cmd/demo/test_server.go

# now for running a second server, in another terminal window, run:
go run cmd/demo/test_server.go --port 8082

# in another terminal window, send some req to the load balancer server. Run it multiple times to see the actual load balancing
curl localhost:8080
```

Currently, the selection **strategy** for everything that is related to `ServerList` is baked into the `ServerList`. So we don't have
an abstraction for a strategy to deal with for load balancing. What's the correct way to abstract this?

## Part 3
One way is that the config for each replica has other information like a name and a weight associated with them. So they can have a URL and some
metadata associated with them(each).