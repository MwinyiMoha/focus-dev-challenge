-- name: CreateOutboundMessage :one
INSERT INTO outbound_messages (campaign_id, customer_id, status, rendered_content, last_error, retry_count)
VALUES (@campaign_id, @customer_id, @status, @rendered_content, @last_error, @retry_count)
RETURNING *;
