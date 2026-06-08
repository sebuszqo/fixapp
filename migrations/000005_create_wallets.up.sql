-- Wallets and transactions for credit system

CREATE TABLE IF NOT EXISTS wallets (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL UNIQUE REFERENCES users(id),
    balance     INTEGER NOT NULL DEFAULT 0 CHECK (balance >= 0),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS wallet_transactions (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    wallet_id       UUID NOT NULL REFERENCES wallets(id),
    type            VARCHAR(10) NOT NULL CHECK (type IN ('credit', 'debit')),
    amount          INTEGER NOT NULL CHECK (amount > 0),
    reason          VARCHAR(50) NOT NULL CHECK (reason IN (
        'initial_bonus', 'admin_top_up', 'package_purchase',
        'lead_accepted', 'lead_refund', 'client_reward'
    )),
    reference_id    UUID,
    description     VARCHAR(500),
    balance_after   INTEGER NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_wallets_user ON wallets(user_id);
CREATE INDEX idx_wallet_tx_wallet ON wallet_transactions(wallet_id);
CREATE INDEX idx_wallet_tx_created ON wallet_transactions(created_at DESC);
CREATE INDEX idx_wallet_tx_reason ON wallet_transactions(reason);
CREATE INDEX idx_wallet_tx_reference ON wallet_transactions(reference_id) WHERE reference_id IS NOT NULL;

-- Updated_at trigger for wallets
CREATE TRIGGER update_wallets_updated_at
    BEFORE UPDATE ON wallets
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

COMMENT ON TABLE wallets IS 'User credit wallets (1 credit = 1 PLN)';
COMMENT ON TABLE wallet_transactions IS 'Immutable log of all credit/debit operations';
COMMENT ON COLUMN wallets.balance IS 'Current balance in credits (cannot go negative)';
COMMENT ON COLUMN wallet_transactions.reason IS 'Why this transaction occurred';
COMMENT ON COLUMN wallet_transactions.reference_id IS 'Related entity (lead_id, job_id, etc.)';
COMMENT ON COLUMN wallet_transactions.balance_after IS 'Wallet balance after this transaction';
