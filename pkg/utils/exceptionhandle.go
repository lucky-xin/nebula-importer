package utils

import (
	"context"
	"errors"
	"github.com/lucky-xin/nebula-importer/pkg/task/ecode"

	"github.com/vesoft-inc/go-pkg/errorx"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type WithErrorMessage func(c *errorx.ErrCode, err error, formatWithArgs ...interface{}) error

type GormErrorWrapper func(err error, formatWithArgs ...interface{}) error

func GormErrorWithLogger(ctx context.Context) func(err error, formatWithArgs ...interface{}) error {
	logger := logx.WithContext(ctx)
	return func(err error, formatWithArgs ...interface{}) error {
		switch {
		case err == nil:
			return nil
		case errors.Is(err, gorm.ErrRecordNotFound):
			return nil
		default:
			logger.Errorf(err.Error(), formatWithArgs...)
			return ecode.WithErrorMessage(ecode.ErrInternalDatabase, err, formatWithArgs...)
		}
	}
}

func ErrMsgWithLogger(ctx context.Context) WithErrorMessage {
	logger := logx.WithContext(ctx)
	return func(c *errorx.ErrCode, err error, formatWithArgs ...interface{}) error {
		logger.Errorf(err.Error(), formatWithArgs...)
		return ecode.WithErrorMessage(c, err, formatWithArgs...)
	}
}
