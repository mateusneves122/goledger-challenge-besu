BESU_DIR     := besu
CONTRACT_DIR := SimpleStorage
APP_DIR      := app
ENV_FILE     := $(CONTRACT_DIR)/.env
ENV_EXAMPLE  := $(CONTRACT_DIR)/.env.example

.PHONY: devnet stop-devnet deploy devnet-deploy \
        app-run app-build app-down app-db-up app-db-down \
        start stop help

# --- Besu network ---
devnet:
	cd $(BESU_DIR) && ./startBesu.sh

stop-devnet:
	cd $(BESU_DIR) && ./stopBesu.sh

deploy:
	@if [ ! -f $(ENV_FILE) ]; then \
		echo "No .env found — copying from .env.example. Edit $(ENV_FILE) to set a custom PRIVATE_KEY."; \
		cp $(ENV_EXAMPLE) $(ENV_FILE); \
	fi
	@set -a && . ./$(ENV_FILE) && set +a && \
		cd $(CONTRACT_DIR) && forge script script/SimpleStorage.s.sol:SimpleStorageScript \
		--rpc-url besu \
		--broadcast

devnet-deploy: devnet deploy

# --- App ---
app-run:
	cd $(APP_DIR) && docker compose up -d

app-build:
	cd $(APP_DIR) && docker compose build

app-down:
	cd $(APP_DIR) && docker compose down

app-db-up:
	cd $(APP_DIR) && docker compose up -d postgres

app-db-down:
	cd $(APP_DIR) && docker compose stop postgres

# --- Full lifecycle ---
start: devnet-deploy app-run

stop: stop-devnet app-down

# --- Help ---
help:
	@echo ""
	@echo "Besu network:"
	@echo "  make devnet          Start the 4-node Besu network"
	@echo "  make stop-devnet     Stop and clean up the Besu network"
	@echo "  make deploy          Deploy the SimpleStorage contract"
	@echo "  make devnet-deploy   Start network and deploy contract"
	@echo ""
	@echo "App:"
	@echo "  make app-run         Start the API and PostgreSQL"
	@echo "  make app-build       Rebuild the API image"
	@echo "  make app-down        Stop and remove app containers"
	@echo "  make app-db-up       Start only PostgreSQL"
	@echo "  make app-db-down     Stop only PostgreSQL"
	@echo ""
	@echo "Full lifecycle:"
	@echo "  make start           devnet-deploy + app-run"
	@echo "  make stop            stop-devnet + app-down"
	@echo ""
	@echo "First-time setup:"
	@echo "  1. make devnet-deploy          (prints the deployed contract address)"
	@echo "  2. set CONTRACT_ADDRESS=<addr> in app/.env"
	@echo "  3. make app-run"
	@echo ""
	@echo "  NOTE: 'make start' assumes CONTRACT_ADDRESS is already set in app/.env."
	@echo ""
