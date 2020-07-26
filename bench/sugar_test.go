package main

import (
	"io/ioutil"
	"testing"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func BenchmarkSugar(b *testing.B) {
	// config := zap.NewProductionConfig()
	// config.OutputPaths = []string{
	// 	"/tmp/bemchmark.log",
	// }
	// logger, _ := config.Build()

	ec := zap.NewProductionEncoderConfig()
	ec.EncodeTime = zapcore.ISO8601TimeEncoder

	file, _ := ioutil.TempFile("", "benchmark.log")

	logger := zap.New(zapcore.NewCore(
		zapcore.NewJSONEncoder(ec),
		zapcore.Lock(file),
		zapcore.InfoLevel,
	))

	sugar := logger.Sugar()

	defer sugar.Sync()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sugar.Info("Benchmark")
	}
}

func BenchmarkSugarFormat(b *testing.B) {
	// config := zap.NewProductionConfig()
	// config.OutputPaths = []string{
	// 	"/tmp/bemchmark.log",
	// }
	// logger, _ := config.Build()

	ec := zap.NewProductionEncoderConfig()
	ec.EncodeTime = zapcore.ISO8601TimeEncoder

	file, _ := ioutil.TempFile("", "benchmark.log")

	logger := zap.New(zapcore.NewCore(
		zapcore.NewJSONEncoder(ec),
		zapcore.Lock(file),
		zapcore.InfoLevel,
	))

	sugar := logger.Sugar()

	defer sugar.Sync()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sugar.Infof("Benchmark %d", i)
	}
}

func BenchmarkSugarDiscardWriter(b *testing.B) {
	ec := zap.NewProductionEncoderConfig()
	// ec.EncodeDuration = zapcore.NanosDurationEncoder
	ec.EncodeTime = zapcore.EpochNanosTimeEncoder
	enc := zapcore.NewJSONEncoder(ec)
	logger := zap.New(zapcore.NewCore(
		enc,
		&Discarder{},
		zapcore.InfoLevel,
	))

	sugar := logger.Sugar()

	defer sugar.Sync()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sugar.Info("Benchmark")
	}
}

func BenchmarkSugarWithoutFlags(b *testing.B) {
	ec := zap.NewProductionEncoderConfig()
	ec.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) { enc.AppendInt(0) }
	// ec.EncodeTime = zapcore.ISO8601TimeEncoder
	file, _ := ioutil.TempFile("", "benchmark.log")

	logger := zap.New(zapcore.NewCore(
		zapcore.NewJSONEncoder(ec),
		zapcore.Lock(file),
		zapcore.InfoLevel,
	))

	sugar := logger.Sugar()

	defer sugar.Sync()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sugar.Infof("Benchmark")
	}
}

func BenchmarkSugarWithDebugLevel(b *testing.B) {
	ec := zap.NewProductionEncoderConfig()
	ec.EncodeTime = zapcore.ISO8601TimeEncoder

	file, _ := ioutil.TempFile("", "benchmark.log")

	logger := zap.New(zapcore.NewCore(
		zapcore.NewJSONEncoder(ec),
		zapcore.Lock(file),
		zapcore.DebugLevel,
	))

	sugar := logger.Sugar()

	defer sugar.Sync()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sugar.Infof("Benchmark")
	}
}

func BenchmarkSugarWithFields(b *testing.B) {
	ec := zap.NewProductionEncoderConfig()
	ec.EncodeTime = zapcore.ISO8601TimeEncoder

	file, _ := ioutil.TempFile("", "benchmark.log")

	logger := zap.New(zapcore.NewCore(
		zapcore.NewJSONEncoder(ec),
		zapcore.Lock(file),
		zapcore.InfoLevel,
	))

	sugar := logger.Sugar()

	defer sugar.Sync()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sugar.Infof("Benchmark", zap.String("url", "www.demo.com"), zap.Int("attempt", 5), zap.Duration("backoff", time.Second))
	}
}

func BenchmarkSugarWithFieldsFormat(b *testing.B) {
	ec := zap.NewProductionEncoderConfig()
	ec.EncodeTime = zapcore.ISO8601TimeEncoder

	file, _ := ioutil.TempFile("", "benchmark.log")

	logger := zap.New(zapcore.NewCore(
		zapcore.NewJSONEncoder(ec),
		zapcore.Lock(file),
		zapcore.InfoLevel,
	))

	sugar := logger.Sugar()

	defer sugar.Sync()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sugar.Infof("Benchmark %d", i, zap.String("url", "www.demo.com"), zap.Int("attempt", 5), zap.Duration("backoff", time.Second))
	}
}
