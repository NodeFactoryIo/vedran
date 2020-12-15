# Vedran

> Polkadot chain load balancer.

### Architecture

_Vedran loadbalancer_ is used in conjunction with [Vedran daemon](https://github.com/NodeFactoryIo/vedran-daemon).
Suppose the node owner wants to register to loadbalancer, than it is required to install and run _Vedran daemon_.
Daemon executes the registration process and starts providing all relevant information (ping, metrics) to the _Vedran loadbalancer_.
Please check [Vedran daemon repo](https://github.com/NodeFactoryIo/vedran-daemon) for more details on the daemon itself.

![Image of vedran architecture](./assets/vedran-arch.png)

## Demo

_This is dockerized demo of entire setup with loadbalancer, node and daemon_

### Requirements

- Install [Docker Engine](https://docs.docker.com/engine/install/)
- Install [Docker Compose](https://docs.docker.com/compose/install/)

**Run demo with `docker-compose up`**

_After all components have been started and node has sent first valid metrics report (after 30 seconds),
you can invoke RPC methods on `localhost:4000` using HTTP requests or on `localhost:4000/ws` using WebSocket request_
Metrics can be seen on `localhost:3000` hosted grafana under `vedran-dashboard`

You can check available rpc methods with:

```bash
curl -H "Content-Type: application/json" -d '{"id":1, "jsonrpc":"2.0", "method": "rpc_methods"}' http://localhost:4000
```
This demo starts five separate dockerized components:
- _Polkadot node_ ([repository](https://github.com/paritytech/polkadot))
- _Vedran daemon_ ([repository](https://github.com/NodeFactoryIo/vedran-daemon))
- _Vedran loadbalancer_ (port: 4000)
- _Prometheus server_ (port: 9090) - scrapes metrics from vedran's `/metrics` endpoint
- _Grafana_ (port: 3000) - Visualizes metrics

### Trigger Manual Payout

Our compose setup runs dev chain and our load balancer uses Allice account to do payout
so you don't have to obtain dev DOTs. Polkadot node operator is Bob (he received payout from Allice).
Load balancer in this setup runs payout daily, if you don't wan't to wait,
you can run following command which will create additional container (in compose network) and trigger payout from Allice account:

```
docker run --network vedran_default nodefactory/vedran:v0.3.0 payout --private-key 0xe5be9a5092b81bca64be81d212e7f2f9eba183bb7a90954f7b76361f6edb5c0a --payout-reward 100 --load-balancer-url "http://vedran:4000/ws"
```

## Get **vedran** binary releases

Download prebuild binary from [releases](https://github.com/NodeFactoryIo/vedran/releases).
Be careful to chose appropriate binary depending on your OS. For more details on how to run _vedran loadbalancer_ see [Starting loadbalancer](#starting-loadbalancer) part.

## Get **vedran** package
Alternatively, it is possible to get _vedran_ golang package:
1. Install [Golang](https://golang.org/doc/install) **1.15 or greater**
2. Run the command below
```
go get github.com/NodeFactoryIo/vedran
```
3. Run vedran from your Go bin directory. For linux systems it will likely be:
```
~/go/bin/vedran
```
Note that if you need to do this, you probably want to add your Go bin directory to your $PATH to make things easier!


## Starting loadbalancer

First download latest prebuilt binaries [from releases](https://github.com/NodeFactoryIo/vedran/releases) and unzip it.

Load balancer is started by invoking **start** command.

For example `./vedran start --auth-secret=supersecret --private-key=lb-wallet-private-key`.

For more information you can always run vedran with `--help` flag.
For list of all commands run `vedran --help` or for list of all options for specific command run `vedran start --help`.

**Load balancer will expose Polkadot RPC API on port 80 by default (can be changed using flag `--server-port`)**

Vedran loadbalancer supports both **HTTP** and **Websockets** protocols for Polkadot RPC API.

- **HTTP - available on root path** `/`
- **WS - available on separate path** `/ws`

**For production use certificates (e.g. https://certbot.eff.org/) should be generated and passsed via flags: `--key-file`, `--cert-file` and port changed to 443**

Start command will start application on 2 ports that need to be exposed to public:
 1. RPC entrypoint to nodes and API for nodes to register to load balancer (default: 80)
 2. HTTP tunnel server for creating tunnels between the node and load balancer so node operators don't to have expose nodes to public network (default: 5223)

### Required flags

`--auth-secret` - authentication secret used for generating tokens

`--private-key` - loadbalancers wallet private key, used for sending founds on payout

### Most important flags

| Flag | Description | Default value |
|----|-----------|:--------:|
|`--server-port`|port on which RPC API is exposed|80|
|`--public-ip`|public IP address of loadbalancer|uses multiple services to find out public IP|
|`--cert-file`|SSL certification file|uses HTTP|
|`--key-file`|SSL price key file|uses HTTP|
|`--tunnel-port`|port on which tunnel server is listening for connect requests|5223|
|`--tunnel-port-range`|range of ports that will be used for creating tunnels|20000:30000|

### Other flags

| Flag | Description | Default value |
|----|-----------|:--------:|
|`--name`|public name for load balancer|autogenerated name is used|
|`--capacity`|maximum number of nodes allowed to connect|unlimited capacity|
|`--whitelist`|comma separated list of node id-s, if provided only these nodes will be allowed to connect. This flag can't be used together with --whitelist-file flag, only one option for setting whitelisted nodes can be used|all nodes are whitelisted|
|`--whitelist-file`|path to file with node id-s in each line, if provided only these nodes will be allowed to connect. This flag can't be used together with --whitelist flag, only one option for setting whitelisted nodes can be used|all nodes are whitelisted|
|`--fee`|value between 0-1 representing fixed fee percentage that loadbalancer will take|0.1 (10%)|
|`--selection`|type of selection that is used for selecting nodes on new request, valid values are `round-robin` and `random`|`round-robin`|
|`--payout-interval`|automatic payout interval specified as number of days, for more details see [payout instructions](#payouts)|-|
|`--payout-reward`|defined reward amount that will be distributed on the payout (amount in Planck), for more details see [payout instructions](#payouts)|-|
|`--log-level`|log level (debug, info, warn, error)|error|
|`--log-file`|path to file in which logs will be saved|`stdout`|

### Obtaining DOTs
If you want to do anything on Polkadot, Kusama, or Westend, then you'll need to get an account and some DOT, KSM, or WND tokens, respectively.
When initializing payout, you will provide loadbalancer with created account and from this account rewards will be sent to connected nodes on payout.

For Westend's WND tokens, see the faucet [instructions on the Wiki](https://wiki.polkadot.network/docs/en/learn-DOT#getting-westies).

## Payouts

### Automatic payout

When starting _vedran loadbalancer_ it is possible to configure automatic payout by providing these flags:

`--private-key` - loadbalancers wallet private key (string representation of hex value prefixed with 0x), used for sending rewards on payout

`--payout-interval` - automatic payout interval specified as number of days

`--payout-reward` - defined total reward amount that will be distributed on the payout (amount in Planck)

If all flags have been provided, then each {_payout-interval_} days automatic payout will be started.

### Manual payout

It is possible to run payout script at any time by invoking `vedran payout` command trough console.
This command has two required flags:

`--private-key` - loadbalancers wallet private key (string representation of hex value prefixed with 0x), used for sending founds on payout

`--payout-reward` - defined total reward amount that will be distributed on the payout (amount in Planck)

Additionally, it is possible to change url on which payout script will connect with loadbalancer when executing transactions by setting flag (default value will be _http://localhost:80_)

`--load-balancer-url` - loadbalancer url

### Get private key
You can use [subkey](https://substrate.dev/docs/en/knowledgebase/integrate/subkey) tool to get private key for your wallet.

After installing subkey tool call `subkey inspect "insert your mnemonic here"`.
You can find private key as _Secreet seed_. See example output of subkey command:

```
  Secret seed:      0x1a84771145cdcee05e49142aaff2e5d669ce4b29344a09b973b751ae661acabf
  Public key (hex): 0xa4548fa9b3b15dc4d1c59789952f0ccf6138dd63faf802637895c941f0522d35
  Account ID:       0xa4548fa9b3b15dc4d1c59789952f0ccf6138dd63faf802637895c941f0522d35
  SS58 Address:     5FnAq6wrMzri5V6jLfKgBkbR2rSAMkVAHVYWa3eU7TAV5rv9
```

## Monitoring

Monitoring is done via grafana and prometheus which are expected to be installed.

### Installation
 - [Grafana installation](https://grafana.com/docs/grafana/latest/installation/)
 - [Prometheus installation](https://prometheus.io/docs/prometheus/latest/installation/)

### Configuration

-  ### Grafana
    Should be configured to fetch data from prometheus server as data source [Tutorial](https://prometheus.io/docs/visualization/grafana/).

    Should have a dashboard that visualizes data scraped from prometheus server. Example configuration can be found [here](./infra/grafana/provisioning/dashboards/vedran-dashboard.json) and can be imported like [this](https://grafana.com/docs/grafana/latest/dashboards/export-import/#importing-a-dashboard).


-  ### Prometheus
    Prometheus should be configured to scrape metrics from vedran's `/metrics` endpoint
    via prometheus .yml configuration. Example of which can be found [here](./infra/prometheus/prometheus.yml)


## Vedran loadbalancer API

`POST   api/v1/nodes`

Register node to loadbalancer. Body should contain details about node:

```json
{
  "id": "string",
  "config_hash": "string",
  "payout_address": "string"
}
```

Returns **token** used for invoking rest of API and **tunnel_server_address** on which daemon can open tunnel toward loadbalancer.

```json
{
  "token": "string",
  "tunnel_server_address": "string"
}
```

---

`POST   api/v1/nodes/pings`

Ping loadbalancer from node. Auth token should be in header as `X-Auth-Header`.

---

`PUT    api/v1/nodes/metrics`

Send metrics for node. Auth token should be in header as `X-Auth-Header`. Body should contain metrics as:

```json
{
  "peer_count": "int32",
  "best_block_height": "int64",
  "finalized_block_height": "int64",
  "ready_transaction_count": "int32"
}
```

---

`GET    api/v1/stats`

Returns statistics for all nodes (mapped on node payout address).

```json
{
  "node_1_payout_address": {
    "total_pings": "float64",
    "total_requests": "float64"
  },
  "node_2_payout_address": {
    "total_pings": "float64",
    "total_requests": "float64"
  },
}
```

## Development

### Clone

```bash
git clone git@github.com:NodeFactoryIo/vedran.git
```

### Lint
[Golangci-lint](https://golangci-lint.run/usage/install/#local-installation) is expected to be installed.

```bash
make lint
```

### Build

```bash
make build
```

### Test

```bash
make test
```

## License

This project is licensed under Apache 2.0:
- Apache License, Version 2.0, ([LICENSE-APACHE](http://www.apache.org/licenses/LICENSE-2.0))
