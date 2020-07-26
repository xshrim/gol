package main

import (
	"io/ioutil"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
)

// type NopFormatter struct {
// }

// func (f *NopFormatter) Format(e *logrus.Entry) ([]byte, error) {
// 	return emptyByte, nil
// }
func BenchmarkLogrus(b *testing.B) {
	log := logrus.New()
	file, _ := ioutil.TempFile("", "benchmark.log")
	log.Out = file

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		log.Info("Benchmark")
	}
}

func BenchmarkLogrusFormat(b *testing.B) {
	log := logrus.New()
	file, _ := ioutil.TempFile("", "benchmark.log")
	log.Out = file

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		log.Infof("Benchmark %d", i)
	}
}

func BenchmarkLogrusWithDiscardWriter(b *testing.B) {
	log := logrus.New()
	log.Out = ioutil.Discard

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		log.Info("Benchmark")
	}
}

func BenchmarkLogrusWithoutFlags(b *testing.B) {
	log := logrus.New()
	file, _ := ioutil.TempFile("", "benchmark.log")
	log.Out = file

	log.SetFormatter(&logrus.TextFormatter{
		DisableColors:          true,
		DisableLevelTruncation: true,
		DisableQuote:           true,
		DisableSorting:         true,
		DisableTimestamp:       true,
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		log.Info("Benchmark")
	}
}

// func BenchmarkLogrusWithNullFormatter(b *testing.B) {
// 	logrus.SetFormatter(&NopFormatter{})
// 	log := logrus.New()
// 	file, err := ioutil.TempFile("", "benchmark-log")
// 	if err != nil {
// 		log.Out = os.Stderr
// 	} else {
// 		log.Out = file
// 	}
// 	b.ResetTimer()
// 	for i := 0; i < b.N; i++ {
// 		log.Info("Benchmark ", i)
// 	}
// }

func BenchmarkLogrusWithDebugLevel(b *testing.B) {
	logrus.SetLevel(logrus.DebugLevel)
	log := logrus.New()
	file, _ := ioutil.TempFile("", "benchmark.log")
	log.Out = file

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		log.Info("Benchmark ", i)
	}
}

func BenchmarkLogrusWithFields(b *testing.B) {
	log := logrus.New()
	file, _ := ioutil.TempFile("", "benchmark.log")
	log.Out = file

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		log.WithFields(logrus.Fields{"url": "www.demo.com", "attemp": 5, "duration": time.Second}).Info("Benchmark")
	}
}

func BenchmarkLogrusWithFieldsFormat(b *testing.B) {
	log := logrus.New()
	file, _ := ioutil.TempFile("", "benchmark.log")
	log.Out = file

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		log.WithFields(logrus.Fields{"url": "www.demo.com", "attemp": 5, "duration": time.Second}).Infof("Benchmark %d", i)
	}
}
