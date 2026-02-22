package domain

import "errors"

// ErrShortURLConflict возвращается при попытке создать URL с уже существующим коротким идентификатором.
var ErrShortURLConflict = errors.New("url already exists")

// ErrOriginalURLConflict возвращается при попытке создать URL с уже существующим исходным URL.
var ErrOriginalURLConflict = errors.New("original url already exists")

// IsErrOriginalURLConflict проверяет, является ли ошибка ErrOriginalURLConflict.
func IsErrOriginalURLConflict(err error) bool {
	return errors.Is(err, ErrOriginalURLConflict)
}
