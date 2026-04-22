package logger

import (
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	log    *zap.Logger
	sugar  *zap.SugaredLogger
	level  zap.AtomicLevel
	AppVer string
)

// Logger wraps zap.SugaredLogger with printf-style compatibility methods
// used in legacy parts of this codebase.
type Logger struct {
	*zap.SugaredLogger
}

// Printf provides compatibility with log.Printf call sites.
func (l *Logger) Printf(template string, args ...interface{}) {
	l.Infof(template, args...)
}

// Println provides compatibility with log.Println call sites.
func (l *Logger) Println(args ...interface{}) {
	l.Info(args...)
}

// Init creates a new logger with the specified configuration
func Init(version, logPath string, debugMode bool) error {
	AppVer = version
	level = zap.NewAtomicLevel()
	if debugMode {
		level.SetLevel(zapcore.DebugLevel)
	}

	// Create log directory if needed
	if logPath != "" {
		if err := os.MkdirAll(filepath.Dir(logPath), 0755); err != nil {
			return err
		}
	}

	// Configure logging format
	config := zap.Config{
		Level:            level,
		Development:      debugMode,
		Encoding:         "console",
		EncoderConfig:    zap.NewDevelopmentEncoderConfig(),
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}

	if logPath != "" {
		config.OutputPaths = append(config.OutputPaths, logPath)
	}

	// Customize time format
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.EncoderConfig.EncodeCaller = zapcore.ShortCallerEncoder

	var err error
	log, err = config.Build(
		zap.AddStacktrace(zapcore.ErrorLevel),
		zap.Fields(zap.String("version", version)),
	)
	if err != nil {
		return err
	}

	sugar = log.Sugar()
	return nil
}

// Get returns a named logger instance
func Get(name string) *Logger {
	if sugar == nil {
		sugar = zap.NewNop().Sugar()
	}
	return &Logger{SugaredLogger: sugar.Named(name)}
}

// SetLevel dynamically changes the logging level
func SetLevel(lvl zapcore.Level) {
	level.SetLevel(lvl)
}
