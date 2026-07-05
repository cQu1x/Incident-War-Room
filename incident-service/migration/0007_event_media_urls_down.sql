-- Restore the single-media columns and carry over the first media URL.

ALTER TABLE incident_events
    ADD COLUMN IF NOT EXISTS media_url TEXT,
    ADD COLUMN IF NOT EXISTS image_url TEXT;

UPDATE incident_events
SET media_url = (media_urls::jsonb ->> 0)
WHERE media_urls IS NOT NULL AND jsonb_array_length(media_urls::jsonb) > 0;

ALTER TABLE incident_events
    DROP COLUMN IF EXISTS media_urls;
