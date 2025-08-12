package logger

import (
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

// Logger はアプリケーション全体で使用するロガーのインスタンス
var Logger *logrus.Logger

// Config はロガーの設定を定義
type Config struct {
	Level  string `json:"level"`
	Format string `json:"format"`
	Output string `json:"output"`
}

// Init はロガーを初期化します
func Init(config Config) error {
	Logger = logrus.New()

	// ログレベルの設定
	level, err := logrus.ParseLevel(config.Level)
	if err != nil {
		level = logrus.InfoLevel
	}
	Logger.SetLevel(level)

	// ログフォーマットの設定
	switch config.Format {
	case "json":
		Logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: time.RFC3339,
			FieldMap: logrus.FieldMap{
				logrus.FieldKeyTime:  "timestamp",
				logrus.FieldKeyLevel: "level",
				logrus.FieldKeyMsg:   "message",
			},
		})
	case "text":
		Logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: time.RFC3339,
		})
	default:
		Logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: time.RFC3339,
		})
	}

	// ログ出力先の設定
	switch config.Output {
	case "stdout":
		Logger.SetOutput(os.Stdout)
	case "stderr":
		Logger.SetOutput(os.Stderr)
	case "file":
		file, err := os.OpenFile("logs/app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return err
		}
		Logger.SetOutput(file)
	default:
		Logger.SetOutput(os.Stdout)
	}

	return nil
}

// GetLogger はロガーインスタンスを取得します
func GetLogger() *logrus.Logger {
	if Logger == nil {
		// デフォルト設定で初期化
		Init(Config{
			Level:  "info",
			Format: "json",
			Output: "stdout",
		})
	}
	return Logger
}

// WithField はフィールド付きのロガーを返します
func WithField(key string, value interface{}) *logrus.Entry {
	return GetLogger().WithField(key, value)
}

// WithFields は複数フィールド付きのロガーを返します
func WithFields(fields logrus.Fields) *logrus.Entry {
	return GetLogger().WithFields(fields)
}

// WithError はエラー付きのロガーを返します
func WithError(err error) *logrus.Entry {
	return GetLogger().WithError(err)
}

// Debug はデバッグレベルのログを出力します
func Debug(args ...interface{}) {
	GetLogger().Debug(args...)
}

// Info は情報レベルのログを出力します
func Info(args ...interface{}) {
	GetLogger().Info(args...)
}

// Warn は警告レベルのログを出力します
func Warn(args ...interface{}) {
	GetLogger().Warn(args...)
}

// Error はエラーレベルのログを出力します
func Error(args ...interface{}) {
	GetLogger().Error(args...)
}

// Fatal は致命的エラーレベルのログを出力し、アプリケーションを終了します
func Fatal(args ...interface{}) {
	GetLogger().Fatal(args...)
}

// Debugf はデバッグレベルのフォーマット付きログを出力します
func Debugf(format string, args ...interface{}) {
	GetLogger().Debugf(format, args...)
}

// Infof は情報レベルのフォーマット付きログを出力します
func Infof(format string, args ...interface{}) {
	GetLogger().Infof(format, args...)
}

// Warnf は警告レベルのフォーマット付きログを出力します
func Warnf(format string, args ...interface{}) {
	GetLogger().Warnf(format, args...)
}

// Errorf はエラーレベルのフォーマット付きログを出力します
func Errorf(format string, args ...interface{}) {
	GetLogger().Errorf(format, args...)
}

// Fatalf は致命的エラーレベルのフォーマット付きログを出力し、アプリケーションを終了します
func Fatalf(format string, args ...interface{}) {
	GetLogger().Fatalf(format, args...)
}
