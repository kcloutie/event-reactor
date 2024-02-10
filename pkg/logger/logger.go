package logger

import (
	"context"
	"fmt"
	"log"
	"os"
	"runtime/debug"
	"strings"
	"sync"

	"github.com/kcloutie/event-reactor/pkg/params/settings"
	"github.com/kcloutie/event-reactor/pkg/params/version"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type ctxKey struct{}

var (
	once   sync.Once
	logger *zap.Logger
)

const (
	RootCommandKey = "root_command"
	SubCommandKey  = "sub_command"
	CommitKey      = "commit"
	VersionKey     = "version"
	BuildTimeKey   = "build_time"
	GoVersionKey   = "go_version"
)

type Level int8

const (
	// DebugLevel logs are typically voluminous, and are usually disabled in
	// production.
	Trace Level = iota - 2
)

// Get initializes a zap.Logger instance if it has not been initialized
// already and returns the same instance for subsequent calls.
func Get() *zap.Logger {
	once.Do(func() {
		stdout := zapcore.AddSync(os.Stdout)

		file := zapcore.AddSync(&lumberjack.Logger{
			Filename:   fmt.Sprintf("%s.log", settings.CliBinaryName),
			MaxSize:    5,
			MaxBackups: 10,
			MaxAge:     14,
			Compress:   true,
		})

		logLevel := getLogLevelFromEnv()

		encodingLoggerConfig := getGcpEncodingConfig()

		erLog := os.Getenv("EVENT_REACTOR_LOG")
		logToConsole := strings.Contains(strings.ToUpper(erLog), "CONSOLE")
		logToConsoleJson := strings.Contains(strings.ToUpper(erLog), "CONSOLEJSON")

		if settings.DebugModeEnabled {
			logToConsole = true
		}

		buildInfo, _ := debug.ReadBuildInfo()
		additionalFields := []zapcore.Field{
			zap.String(CommitKey, version.Commit),
			zap.String(VersionKey, version.BuildVersion),
			zap.String(BuildTimeKey, version.BuildTime),
			zap.String(GoVersionKey, buildInfo.GoVersion),
		}

		if logToConsoleJson {
			// when logging to console in json format, we will not be writing to a log file
			logger = createJsonConsoleLogger(*encodingLoggerConfig, logLevel, buildInfo)
			return
		} else {
			encodingLoggerConfig.EncodeLevel = zapcore.CapitalLevelEncoder
		}

		logToFile := strings.Contains(strings.ToUpper(erLog), "FILE")

		consoleEncoder := zapcore.NewConsoleEncoder(*encodingLoggerConfig)
		if logToConsoleJson {
			consoleEncoder = zapcore.NewJSONEncoder(*encodingLoggerConfig)
		}
		fileEncoder := zapcore.NewJSONEncoder(*encodingLoggerConfig)

		var core zapcore.Core
		if !logToConsole && !logToFile {
			core = zapcore.NewTee()
		}

		if logToConsole && logToFile {

			core = zapcore.NewTee(
				zapcore.NewCore(consoleEncoder, stdout, logLevel),
				zapcore.NewCore(fileEncoder, file, logLevel).
					With(
						additionalFields,
					),
			)
		} else {
			if logToFile {
				core = zapcore.NewTee(
					zapcore.NewCore(fileEncoder, file, logLevel).
						With(
							additionalFields,
						),
				)
			}
			if logToConsole {
				core = zapcore.NewTee(
					zapcore.NewCore(consoleEncoder, stdout, logLevel),
				)
			}
		}

		logger = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.FatalLevel))
	})

	return logger
}

func createJsonConsoleLogger(encoderCfg zapcore.EncoderConfig, logLevel zap.AtomicLevel, buildInfo *debug.BuildInfo) *zap.Logger {
	config := zap.Config{
		Level:             logLevel,
		Development:       false,
		DisableCaller:     false,
		DisableStacktrace: false,
		Sampling:          nil,
		Encoding:          "json",
		EncoderConfig:     encoderCfg,
		OutputPaths: []string{
			"stdout",
		},
		ErrorOutputPaths: []string{
			"stderr",
		},
		InitialFields: map[string]interface{}{
			CommitKey:    version.Commit,
			VersionKey:   version.BuildVersion,
			BuildTimeKey: version.BuildTime,
			GoVersionKey: buildInfo.GoVersion,
		},
	}

	return zap.Must(config.Build())
}

func getLogLevelFromEnv() zap.AtomicLevel {
	level := zap.InfoLevel
	levelEnv := os.Getenv("LOG_LEVEL")
	if levelEnv != "" {
		levelFromEnv, err := zapcore.ParseLevel(levelEnv)
		if err != nil {
			log.Println(
				fmt.Errorf("invalid level, defaulting to INFO: %w", err),
			)
		}

		level = levelFromEnv
	} else {
		if settings.DebugModeEnabled {
			level = zap.DebugLevel
		}
	}

	logLevel := zap.NewAtomicLevelAt(level)
	fmt.Println("logLevel: ", logLevel.String())
	return logLevel
}

func getGcpEncodingConfig() *zapcore.EncoderConfig {
	config := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "severity",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    gcpEncodeLevel(),
		EncodeTime:     zapcore.RFC3339TimeEncoder,
		EncodeDuration: zapcore.MillisDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
	return &config
}

func gcpEncodeLevel() zapcore.LevelEncoder {
	return func(l zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
		switch l {
		case zapcore.DebugLevel:
			enc.AppendString("DEBUG")
		case zapcore.InfoLevel:
			enc.AppendString("INFO")
		case zapcore.WarnLevel:
			enc.AppendString("WARNING")
		case zapcore.ErrorLevel:
			enc.AppendString("ERROR")
		case zapcore.DPanicLevel:
			enc.AppendString("CRITICAL")
		case zapcore.PanicLevel:
			enc.AppendString("ALERT")
		case zapcore.FatalLevel:
			enc.AppendString("EMERGENCY")
		}
	}
}

// FromCtx returns the Logger associated with the ctx. If no logger
// is associated, the default logger is returned, unless it is nil
// in which case a disabled logger is returned.
func FromCtx(ctx context.Context) *zap.Logger {
	l, exists := ctx.Value(ctxKey{}).(*zap.Logger)
	if exists {
		return l
	} else {
		l := logger
		if l != nil {
			return l
		}
	}
	return zap.NewNop()
}

func FromCtxWithCtx(ctx context.Context, fields ...zapcore.Field) (*zap.Logger, context.Context) {
	log := FromCtx(ctx)
	log = log.With(fields...)
	ctx = WithCtx(ctx, log)
	return log, ctx
}

// WithCtx returns a copy of ctx with the Logger attached.
func WithCtx(ctx context.Context, l *zap.Logger) context.Context {
	lp, exists := ctx.Value(ctxKey{}).(*zap.Logger)
	if exists {
		if lp == l {
			return ctx
		}
	}
	return context.WithValue(ctx, ctxKey{}, l)
}

type LeveledLogger struct {
	logger *zap.Logger
}

func NewLeveledLogger(lgr *zap.Logger) LeveledLogger {
	l := LeveledLogger{logger: lgr.WithOptions(zap.AddCallerSkip(3))}
	return l
}

func (l *LeveledLogger) Error(msg string, keysAndValues ...interface{}) {
	lgr := addFields(l.logger, keysAndValues)
	lgr.Level()
	lgr.Error(msg)
}

func (l *LeveledLogger) Info(msg string, keysAndValues ...interface{}) {
	lgr := addFields(l.logger, keysAndValues)
	lgr.Info(msg)
}

func (l *LeveledLogger) Debug(msg string, keysAndValues ...interface{}) {
	lgr := addFields(l.logger, keysAndValues)
	lgr.Debug(msg)
}

func (l *LeveledLogger) Warn(msg string, keysAndValues ...interface{}) {
	lgr := addFields(l.logger, keysAndValues)
	lgr.Warn(msg)
}

func addFields(lgr *zap.Logger, keysAndValues []interface{}) *zap.Logger {
	for i := 0; i < len(keysAndValues)-1; i += 2 {
		lgr = lgr.With(zap.String(keysAndValues[i].(string), fmt.Sprintf("%v", keysAndValues[i+1])))
	}
	return lgr
}
