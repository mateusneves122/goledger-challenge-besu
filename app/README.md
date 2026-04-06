# SimpleStorage App

REST API in Go that interacts with the `SimpleStorage` smart contract on Hyperledger Besu and syncs its value to PostgreSQL.

## First-time setup

All commands below should be run from the **repository root**.

```bash
# 1. Start the Besu network and deploy the contract (the address is printed at the end)
make devnet-deploy

# 2. Set the deployed address in app/.env
#    Copy the example and fill in CONTRACT_ADDRESS
cp app/.env.example app/.env

# 3. Start the API and PostgreSQL
make app-run
```

> For a full list of available commands, run `make help` from the repository root.

## Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/api/storage` | Set a new value in the smart contract |
| `GET` | `/api/storage` | Get the current value from the smart contract |
| `POST` | `/api/storage/sync` | Sync the blockchain value to the database |
| `GET` | `/api/storage/check` | Check if blockchain and database values match |

**Set** — sends a new value to the smart contract:
```bash
curl -X POST localhost:8080/api/storage \
  -H 'Content-Type: application/json' \
  -d '{"value":"42"}'
# {"tx_hash":"0x...","value":"10000000000000000000000000000000011"}
```

**Get** — reads the current value from the blockchain:
```bash
curl localhost:8080/api/storage
# {"value":"10000000000000000000000000000000011"}
```

**Sync** — persists the blockchain value to the database:
```bash
curl -X POST localhost:8080/api/storage/sync
# {"message":"synced","value":"10000000000000000000000000000000011"}
```

**Check** — compares blockchain and database values:
```bash
curl localhost:8080/api/storage/check
# {"match":true}
```

## Architecture

The application follows a layered architecture with clear separation of concerns. Each layer depends only on the layer below it through interfaces, making the business logic independent of infrastructure (blockchain or database). Dependency injection is done manually at startup in `cmd/server.go`, where the adapters are wired into the service and the service into the handler.

Layered structure inside `internal/simplestorage/`:

- **handler** — receives HTTP requests, validates input (required field, non-negative integer), delegates to the service, and returns JSON responses with appropriate status codes
- **service** — orchestrates the four use cases (Set, Get, Sync, Check); depends only on interfaces (`ContractReader`, `ContractWriter`, `Repository`), not on concrete implementations
- **blockchain** — implements the contract interfaces using `go-ethereum/bind`; read operations call `eth_call` (no gas, no signing), write operations sign transactions with the configured private key and wait for the receipt via `bind.WaitMined`
- **repository** — implements the `Repository` interface using `pgx/v5`; persists the synced value per contract address via upsert, and returns `"0"` when no record exists yet

The contract ABI is embedded at compile time via `//go:embed`, removing any runtime file dependency.

## Environment variables

| Variable | Description |
|----------|-------------|
| `BESU_RPC_URL` | Besu node HTTP RPC endpoint |
| `CONTRACT_ADDRESS` | Deployed contract address (required) |
| `PRIVATE_KEY` | Hex private key for signing transactions |
| `DB_HOST` / `DB_PORT` / `DB_USER` / `DB_PASS` / `DB_NAME` | PostgreSQL connection |
| `SERVER_ADDR` | HTTP listen address (e.g. `:8080`) |
