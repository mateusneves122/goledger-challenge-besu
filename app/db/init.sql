CREATE TABLE IF NOT EXISTS contract_values (
    contract_address VARCHAR(42)  PRIMARY KEY,
    value            NUMERIC      NOT NULL,
    synced_at        TIMESTAMP    NOT NULL DEFAULT NOW()
);
