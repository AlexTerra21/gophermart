package logger

import (
	"net/http"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type LevelLog int16

const (
	_ LevelLog = iota - 1
	LevelDebug
	LevelInfo
	LevelWarning
	LevelError
)

type Field struct {
	Key string
	Val any
}

var _log *zap.Logger = zap.NewNop()

// Log будет доступен всему коду как синглтон.
func Log() *zap.Logger {
	return _log
}

// Initialize инициализирует синглтон логера с необходимым уровнем логирования.
func Initialize(level string) error {
	// преобразуем текстовый уровень логирования в zap.AtomicLevel
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return err
	}

	cfg := zap.NewDevelopmentConfig()
	cfg.Level = lvl

	tmpLog, err := cfg.Build()
	if err != nil {
		return err
	}

	_log = tmpLog
	return nil
}

func log(lvl LevelLog, msg string, vars ...Field) {
	defer func() {
		_ = _log.Sync()
	}()

	s := make([]zap.Field, 0, len(vars))
	for _, val := range vars {
		switch v := val.Val.(type) {
		case zapcore.ObjectMarshaler:
			s = append(s, zap.Object(val.Key, v))
		case string:
			s = append(s, zap.String(val.Key, v))
		case error:
			s = append(s, zap.String(val.Key, v.Error()))
		default:
			s = append(s, zap.Reflect(val.Key, v))
		}
	}

	switch lvl {
	case LevelDebug:
		_log.Debug(msg, s...)
	case LevelInfo:
		_log.Info(msg, s...)
	case LevelWarning:
		_log.Warn(msg, s...)
	case LevelError:
		_log.Error(msg, s...)
	}
}

func Debug(msg string, vars ...Field) {
	log(LevelDebug, msg, vars...)
}
func Info(msg string, vars ...Field) {
	log(LevelInfo, msg, vars...)
}
func Warn(msg string, vars ...Field) {
	log(LevelWarning, msg, vars...)
}
func Error(msg string, vars ...Field) {
	log(LevelError, msg, vars...)
}

func WithLogging(h http.Handler) http.HandlerFunc {
	logFn := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		responseData := &responseData{
			status: 0,
			size:   0,
		}

		lw := loggingResponseWriter{
			ResponseWriter: w, // встраиваем оригинальный http.ResponseWriter
			responseData:   responseData,
		}
		h.ServeHTTP(&lw, r)
		duration := time.Since(start)

		_log.Info("Request",
			zap.String("uri", r.RequestURI),
			zap.String("method", r.Method),
			zap.Int("status", responseData.status),
			zap.Int("size", responseData.size),
			zap.String("duration", duration.String()),
		)
	}

	return logFn
}

type (
	// берём структуру для хранения сведений об ответе
	responseData struct {
		status int
		size   int
	}

	// добавляем реализацию http.ResponseWriter
	loggingResponseWriter struct {
		http.ResponseWriter // встраиваем оригинальный http.ResponseWriter
		responseData        *responseData
	}
)

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	// записываем ответ, используя оригинальный http.ResponseWriter
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size // захватываем размер
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	// записываем код статуса, используя оригинальный http.ResponseWriter
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode // захватываем код статуса
}
