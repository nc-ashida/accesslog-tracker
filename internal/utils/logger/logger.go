package logger

import (
	"errors"
	"io"
	"os"

	"github.com/sirupsen/logrus"
)

// Logger はログ機能を提供するインターフェースです
type Logger interface {
	Debug(args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
	Fatal(args ...interface{})
	Panic(args ...interface{})
	
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})
	Panicf(format string, args ...interface{})
	
	WithField(key string, value interface{}) Logger
	WithFields(fields map[string]interface{}) Logger
	WithError(err error) Logger
	
	SetLevel(level string) error
	SetFormat(format string) error
	SetOutput(output io.Writer)
}

// logger はLoggerインターフェースの実装です
type logger struct {
	entry *logrus.Entry
}

// NewLogger は新しいロガーインスタンスを作成します
func NewLogger() Logger {
	l := logrus.New()
	l.SetOutput(os.Stdout)
	l.SetLevel(logrus.InfoLevel)
	l.SetFormatter(&logrus.JSONFormatter{})
	
	return &logger{entry: logrus.NewEntry(l)}
}

// SetLevel はログレベルを設定します
func (l *logger) SetLevel(level string) error {
	switch level {
	case "debug":
		l.entry.Logger.SetLevel(logrus.DebugLevel)
	case "info":
		l.entry.Logger.SetLevel(logrus.InfoLevel)
	case "warn":
		l.entry.Logger.SetLevel(logrus.WarnLevel)
	case "error":
		l.entry.Logger.SetLevel(logrus.ErrorLevel)
	case "fatal":
		l.entry.Logger.SetLevel(logrus.FatalLevel)
	case "panic":
		l.entry.Logger.SetLevel(logrus.PanicLevel)
	default:
		return errors.New("invalid log level")
	}
	return nil
}

// SetFormat はログフォーマットを設定します
func (l *logger) SetFormat(format string) error {
	switch format {
	case "json":
		l.entry.Logger.SetFormatter(&logrus.JSONFormatter{})
	case "text":
		l.entry.Logger.SetFormatter(&logrus.TextFormatter{})
	default:
		return errors.New("invalid log format")
	}
	return nil
}

// SetOutput はログ出力先を設定します
func (l *logger) SetOutput(output io.Writer) {
	l.entry.Logger.SetOutput(output)
}

// WithField は単一のフィールドを追加したロガーを返します
func (l *logger) WithField(key string, value interface{}) Logger {
	return &logger{entry: l.entry.WithField(key, value)}
}

// WithFields は複数のフィールドを追加したロガーを返します
func (l *logger) WithFields(fields map[string]interface{}) Logger {
	return &logger{entry: l.entry.WithFields(logrus.Fields(fields))}
}

// WithError はエラーを追加したロガーを返します
func (l *logger) WithError(err error) Logger {
	return &logger{entry: l.entry.WithError(err)}
}

// Debug はデバッグレベルのログを出力します
func (l *logger) Debug(args ...interface{}) {
	l.entry.Debug(args...)
}

// Info は情報レベルのログを出力します
func (l *logger) Info(args ...interface{}) {
	l.entry.Info(args...)
}

// Warn は警告レベルのログを出力します
func (l *logger) Warn(args ...interface{}) {
	l.entry.Warn(args...)
}

// Error はエラーレベルのログを出力します
func (l *logger) Error(args ...interface{}) {
	l.entry.Error(args...)
}

// Fatal は致命的エラーレベルのログを出力し、アプリケーションを終了します
func (l *logger) Fatal(args ...interface{}) {
	l.entry.Fatal(args...)
}

// Panic はパニックレベルのログを出力し、パニックを発生させます
func (l *logger) Panic(args ...interface{}) {
	l.entry.Panic(args...)
}

// Debugf はデバッグレベルのフォーマット付きログを出力します
func (l *logger) Debugf(format string, args ...interface{}) {
	l.entry.Debugf(format, args...)
}

// Infof は情報レベルのフォーマット付きログを出力します
func (l *logger) Infof(format string, args ...interface{}) {
	l.entry.Infof(format, args...)
}

// Warnf は警告レベルのフォーマット付きログを出力します
func (l *logger) Warnf(format string, args ...interface{}) {
	l.entry.Warnf(format, args...)
}

// Errorf はエラーレベルのフォーマット付きログを出力します
func (l *logger) Errorf(format string, args ...interface{}) {
	l.entry.Errorf(format, args...)
}

// Fatalf は致命的エラーレベルのフォーマット付きログを出力し、アプリケーションを終了します
func (l *logger) Fatalf(format string, args ...interface{}) {
	l.entry.Fatalf(format, args...)
}

// Panicf はパニックレベルのフォーマット付きログを出力し、パニックを発生させます
func (l *logger) Panicf(format string, args ...interface{}) {
	l.entry.Panicf(format, args...)
}
