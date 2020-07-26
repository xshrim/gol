package main

import (
	"io/ioutil"
	"testing"
	"time"

	"github.com/rs/zerolog"
)

func BenchmarkZerolog(b *testing.B) {
	//zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	file, _ := ioutil.TempFile("", "benchmark.log")
	logger := zerolog.New(file).With().Timestamp().Logger()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info().Msg("Benchmark")
	}
}

func BenchmarkZerologFormat(b *testing.B) {
	//zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	file, _ := ioutil.TempFile("", "benchmark.log")
	logger := zerolog.New(file).With().Timestamp().Logger()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info().Msgf("Benchmark %d", i)
	}
}

func BenchmarkZerologDiscardWriter(b *testing.B) {
	logger := zerolog.New(ioutil.Discard).With().Timestamp().Logger()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info().Msg("Benchmark")
	}
}

func BenchmarkZerologWithoutFlags(b *testing.B) {
	file, _ := ioutil.TempFile("", "benchmark.log")
	logger := zerolog.New(file)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info().Msg("Benchmark")
	}
}

func BenchmarkZerologWithDebugLevel(b *testing.B) {
	file, _ := ioutil.TempFile("", "benchmark.log")
	logger := zerolog.New(file).With().Timestamp().Logger().Level(zerolog.DebugLevel)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info().Msg("Benchmark")
	}
}

func BenchmarkZerologWithFields(b *testing.B) {
	file, _ := ioutil.TempFile("", "benchmark.log")
	logger := zerolog.New(file).With().Timestamp().Logger()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info().Str("url", "www.demo.com").Int("attempt", 5).Dur("duration", time.Second).Msg("Benchmark")
	}
}

func BenchmarkZerologWithFieldsFormat(b *testing.B) {
	file, _ := ioutil.TempFile("", "benchmark.log")
	logger := zerolog.New(file).With().Timestamp().Logger()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info().Str("url", "www.demo.com").Int("attempt", 5).Dur("duration", time.Second).Msgf("Benchmark %d", i)
	}
}
