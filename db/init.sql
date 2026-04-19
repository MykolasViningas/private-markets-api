CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS citext;

CREATE TABLE fund_statuses (
    id SMALLINT PRIMARY KEY,
    name TEXT UNIQUE NOT NULL
);

INSERT INTO fund_statuses (id, name)
VALUES
(1, 'fundraising'),
(2, 'investing'),
(3, 'closed');

CREATE TABLE investor_types (
    id SMALLINT PRIMARY KEY,
    name TEXT UNIQUE NOT NULL
);

INSERT INTO investor_types (id, name)
VALUES
(1, 'individual'),
(2, 'institutional'),
(3, 'family_office');

CREATE TABLE funds (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name TEXT NOT NULL UNIQUE,
    vintage_year INTEGER NOT NULL,
    target_size_usd NUMERIC(19,2) NOT NULL CHECK (target_size_usd > 0),
    status_id SMALLINT NOT NULL REFERENCES fund_statuses(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE investors (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name TEXT NOT NULL,
    investor_type_id SMALLINT NOT NULL REFERENCES investor_types(id),
    email CITEXT NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE investments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    fund_id UUID NOT NULL REFERENCES funds(id) ON DELETE CASCADE,
    investor_id UUID NOT NULL REFERENCES investors(id) ON DELETE CASCADE,
    amount_usd NUMERIC(19,2) NOT NULL CHECK (amount_usd > 0),
    investment_date DATE NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_investments_fund_date ON investments(fund_id, investment_date);
