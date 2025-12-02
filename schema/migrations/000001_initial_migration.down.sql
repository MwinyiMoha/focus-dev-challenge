
-- Drop indexes for outbound_messages
DROP INDEX IF EXISTS idx_outbound_messages_status;
DROP INDEX IF EXISTS idx_outbound_messages_customer_id;
DROP INDEX IF EXISTS idx_outbound_messages_campaign_id;

-- Drop outbound_messages table
DROP TABLE IF EXISTS outbound_messages;


-- Drop indexes for campaigns
DROP INDEX IF EXISTS idx_campaigns_status;
DROP INDEX IF EXISTS idx_campaigns_channel;

-- Drop campaigns table
DROP TABLE IF EXISTS campaigns;

-- Drop indexes for customers
DROP INDEX IF EXISTS idx_customers_phone;

-- Drop customers table
DROP TABLE IF EXISTS customers;
