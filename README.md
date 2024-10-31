# Picasso 
Cosmos L1 blockchain with IBC-v7 and wasm client enabled powering IBC extensions to Solana and Ethereum.

## Hardware Recommendation

* Quad core or larger amd64 CPU
* 64GB+ RAM
* 1TB+ NVMe Storage

## Quick start

Requires [Go 1.20](https://go.dev/doc/install) or higher.

```bash
make install
picad version
```
Then you can run a node with a single command.

```bash
./scripts/testnode.sh   
```

If you have Docker installed, then you can run a local node with a single command.

```bash
docker compose up -d
```
and remove all node running with

```bash
docker compose down
```
