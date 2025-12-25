-- Добавляем поле is_deleted типа BOOLEAN со значением по умолчанию FALSE
ALTER TABLE urls ADD COLUMN is_deleted BOOLEAN NOT NULL DEFAULT FALSE;