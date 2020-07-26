package main

import (
	"io/ioutil"
	"log"
	"testing"
)

// type NopLogger struct {
// 	*log.Logger
// }

// func (l *NopLogger) Println(v ...interface{}) {
// 	// noop
// }

// var nop *NopLogger = &NopLogger{
// 	log.New(os.Stderr, "", log.LstdFlags),
// }

func BenchmarkLog(b *testing.B) {
	log.SetFlags(log.LstdFlags)
	file, _ := ioutil.TempFile("", "benchmark.log")
	log.SetOutput(file)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		log.Print("Benchmark")
	}
}

func BenchmarkLogFormat(b *testing.B) {
	log.SetFlags(log.LstdFlags)
	file, _ := ioutil.TempFile("", "benchmark.log")
	log.SetOutput(file)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		log.Printf("Benchmark %d", i)
	}
}

func BenchmarkLogDiscardWriter(b *testing.B) {
	log.SetFlags(log.LstdFlags)
	log.SetOutput(ioutil.Discard)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		log.Print("Benchmark")
	}
}

func BenchmarkLogWithoutFlags(b *testing.B) {
	log.SetFlags(0)
	file, _ := ioutil.TempFile("", "benchmark.log")
	log.SetOutput(file)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		log.Print("Benchmark")
	}
}

// func BenchmarkLogCustomNullWriter(b *testing.B) {
// 	log.SetFlags(log.LstdFlags)
// 	log.SetOutput(new(NullWriter))
// 	b.ResetTimer()
// 	for i := 0; i < b.N; i++ {
// 		log.Printf("Benchmark %d", i)
// 	}
// }
