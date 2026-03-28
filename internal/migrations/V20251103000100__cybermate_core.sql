-- profiles
CREATE TABLE IF NOT EXISTS profiles (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    telegram_id TEXT NOT NULL UNIQUE,
    avatar TEXT,
    location TEXT,
    role TEXT,
    description TEXT,
    telegram_init_data TEXT,
    username TEXT UNIQUE,
    verified BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- wallets
CREATE TABLE IF NOT EXISTS wallets (
    id BIGSERIAL PRIMARY KEY,
    profile_id BIGINT NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
    balance BIGINT NOT NULL DEFAULT 0,
    total_earned BIGINT NOT NULL DEFAULT 0,
    balance_available BIGINT NOT NULL DEFAULT 0
);

-- wallet transactions
CREATE TABLE IF NOT EXISTS wallet_transactions (
    id BIGSERIAL PRIMARY KEY,
    wallet_id BIGINT NOT NULL REFERENCES wallets(id) ON DELETE CASCADE,
    date TIMESTAMPTZ NOT NULL DEFAULT now(),
    type TEXT NOT NULL CHECK (type IN ('withdrawal','referral','deposit','receive')),
    amount BIGINT NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('pending','approved','rejected','completed')) DEFAULT 'pending',
    description TEXT,
    details TEXT
);

-- withdraw methods
CREATE TABLE IF NOT EXISTS withdraw_methods (
    id BIGSERIAL PRIMARY KEY,
    wallet_id BIGINT NOT NULL REFERENCES wallets(id) ON DELETE CASCADE,
    type TEXT NOT NULL CHECK (type IN ('crypto_wallet','bank_account')),
    details TEXT NOT NULL
);

-- referrals link
CREATE TABLE IF NOT EXISTS referrals (
    id BIGSERIAL PRIMARY KEY,
    referrer_profile_id BIGINT NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
    referee_profile_id BIGINT NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
    completed_tasks_count BIGINT NOT NULL DEFAULT 0,
    earnings BIGINT NOT NULL DEFAULT 0,
    UNIQUE(referrer_profile_id, referee_profile_id)
);


