-- Customers table 

CREATE TABLE customers (
    id                  BIGSERIAL PRIMARY KEY,
    phone               VARCHAR(32) NOT NULL,
    first_name          VARCHAR(100),
    last_name           VARCHAR(100),
    location            VARCHAR(255),
    preferred_product   VARCHAR(255),
    created_at          TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_customers_phone ON customers(phone);

-- Campaigns table 

CREATE TABLE campaigns (
    id              BIGSERIAL PRIMARY KEY,
    name            VARCHAR(255) NOT NULL,
    channel         VARCHAR(20) NOT NULL CHECK (channel IN ('sms', 'whatsapp')),
    status          VARCHAR(20) NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'scheduled', 'sending', 'sent', 'failed')),
    base_template   TEXT NOT NULL,
    scheduled_at    TIMESTAMP NULL,
    created_at      TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_campaigns_channel ON campaigns(channel);
CREATE INDEX idx_campaigns_status ON campaigns(status);

-- Outbound messages table 

CREATE TABLE outbound_messages (
    id               BIGSERIAL PRIMARY KEY,
    campaign_id      BIGINT NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
    customer_id      BIGINT NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    status           VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'sent', 'failed')),
    rendered_content TEXT NOT NULL,
    last_error       TEXT,
    retry_count      INT NOT NULL DEFAULT 0,
    created_at       TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_outbound_messages_campaign_id ON outbound_messages(campaign_id);
CREATE INDEX idx_outbound_messages_customer_id ON outbound_messages(customer_id);
CREATE INDEX idx_outbound_messages_status ON outbound_messages(status);
