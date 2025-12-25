-- Добавляем поле user_id типа UUID
ALTER TABLE urls ADD COLUMN user_id UUID;

-- Создаем индекс для ускорения поиска по user_id
CREATE INDEX IF NOT EXISTS idx_urls_user_id ON urls (user_id);