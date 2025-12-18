package domain

import "errors"

var ErrShortURLConflict = errors.New("url already exists")

var ErrOriginalURLConflict = errors.New("original url already exists")

func IsErrOriginalURLConflict(err error) bool {
	return errors.Is(err, ErrOriginalURLConflict)
}
