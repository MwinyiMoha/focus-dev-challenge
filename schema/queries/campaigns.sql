-- name: CreateCampaign :one
INSERT INTO campaigns (name, channel, status, base_template, scheduled_at)
VALUES (@name, @channel, @status, @base_template, @scheduled_at) 
RETURNING *;

-- name: ListCampaigns :many
SELECT
    c.*,
    COUNT(*) OVER() AS total_count
FROM campaigns c
WHERE
    (@status::text IS NULL OR c.status = @status)
    AND (@channel::text IS NULL OR c.channel = @channel)
ORDER BY id DESC
LIMIT @page_size
OFFSET ((@page_number - 1) * @page_size);

-- name: GetCampaign :one
SELECT
    c.*,
    jsonb_build_object(
        'total_messages', COALESCE(COUNT(om.id), 0),
        'pending',        COALESCE(SUM(CASE WHEN om.status = 'pending' THEN 1 ELSE 0 END), 0),
        'sent',           COALESCE(SUM(CASE WHEN om.status = 'sent' THEN 1 ELSE 0 END), 0),
        'failed',         COALESCE(SUM(CASE WHEN om.status = 'failed' THEN 1 ELSE 0 END), 0)
    ) AS stats
FROM campaigns c
LEFT JOIN outbound_messages om ON om.campaign_id = c.id
WHERE c.id = @campaign_id
GROUP BY c.id;
