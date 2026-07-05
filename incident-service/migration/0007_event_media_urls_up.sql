-- =====================
-- EVENT MEDIA URLS
-- =====================
--
-- Replace the single-media columns (media_url, and the never-used image_url)
-- with a single media_urls column that holds any number of media URLs as a
-- JSON array of strings, so one timeline event can carry several attachments of
-- any media type.

ALTER TABLE incident_events
    ADD COLUMN IF NOT EXISTS media_urls TEXT;

-- Carry over existing single media URLs as one-element JSON arrays.
UPDATE incident_events
SET media_urls = to_jsonb(ARRAY[media_url])::text
WHERE media_url IS NOT NULL AND media_urls IS NULL;

ALTER TABLE incident_events
    DROP COLUMN IF EXISTS media_url,
    DROP COLUMN IF EXISTS image_url;
