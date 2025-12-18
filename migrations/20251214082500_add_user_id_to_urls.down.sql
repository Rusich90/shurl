-- Удаляем индекс
DROP INDEX IF EXISTS idx_urls_user_id;

-- Удаляем поле user_id
ALTER TABLE urls DROP COLUMN user_id;