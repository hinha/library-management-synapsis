package logger

import (
	"context"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm/logger"
	"time"
)

type ZerologGormLogger struct {
	LogLevel logger.LogLevel
	Debug    bool
}

func (z *ZerologGormLogger) LogMode(level logger.LogLevel) logger.Interface {
	z.LogLevel = level
	return z
}

func (z *ZerologGormLogger) Info(ctx context.Context, msg string, args ...interface{}) {
	if z.Debug {
		log.Ctx(ctx).Info().Msgf(msg, args...)
	}
}

func (z *ZerologGormLogger) Warn(ctx context.Context, msg string, args ...interface{}) {
	log.Ctx(ctx).Warn().Msgf(msg, args...)
}

func (z *ZerologGormLogger) Error(ctx context.Context, msg string, args ...interface{}) {
	log.Ctx(ctx).Error().Msgf(msg, args...)
}

func (z *ZerologGormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	sql, rows := fc()
	elapsed := time.Since(begin)

	if !z.Debug && err == nil {
		return // skip if not debug mode and no error
	}

	event := log.Ctx(ctx).Debug()
	if err != nil {
		event = log.Ctx(ctx).Error().Err(err)
	}

	event.
		Dur("elapsed", elapsed).
		Int64("rows", rows).
		Str("sql", sql).
		Msg("gorm trace")
}
