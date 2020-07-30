package main

import (
	"io/ioutil"
	"testing"
	"time"

	"github.com/xshrim/gol"
)

// 常规
// 丢弃输出
// 格式化输出
// 带level
// 10个域上下文

func BenchmarkGol(b *testing.B) {
	file, _ := ioutil.TempFile("", "benchmark.log")
	gol.Flag(gol.Ldefault).Writer(file)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gol.Info("Benchmark")
	}
}

func BenchmarkGolFormat(b *testing.B) {
	file, _ := ioutil.TempFile("", "benchmark.log")
	gol.Flag(gol.Ldefault).Writer(file)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gol.Infof("Benchmark %d", i)
	}
}

func BenchmarkGolDiscardWriter(b *testing.B) {
	gol.Flag(gol.Ldefault).Writer(ioutil.Discard)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gol.Info("Benchmark")
	}
}

func BenchmarkGolWithoutFlags(b *testing.B) {
	file, _ := ioutil.TempFile("", "benchmark.log")
	gol.Flag(0).Writer(file)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gol.Info("Benchmark")
	}
}

func BenchmarkGolWithDebugLevel(b *testing.B) {
	file, _ := ioutil.TempFile("", "benchmark.log")
	gol.Flag(gol.Ldefault).Level(gol.DEBUG).Writer(file)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gol.Info("Benchmark")
	}
}

func BenchmarkGolWithFields(b *testing.B) {
	file, _ := ioutil.TempFile("", "benchmark.log")
	gol.Flag(gol.Ldefault).Writer(file)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gol.With(map[string]interface{}{"url": "www.demo.com", "attempt": 5, "duration": time.Second}).Info("Benchmark")
	}
}

func BenchmarkGolWithFieldsFormat(b *testing.B) {
	file, _ := ioutil.TempFile("", "benchmark.log")
	gol.Flag(gol.Ldefault).Writer(file)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		//gol.With(gol.F{"url": "www.demo.com", "attempt": 5, "duration": time.Second}).Infof("Benchmark %d", i)
		gol.With(nil).Str("url", "www.demo.com").Int("attempt", 5).Dur("duration", time.Second).Infof("Benckmark %d", i)
	}
}
