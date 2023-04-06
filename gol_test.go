package gol

import (
	"testing"
	"time"
)

func TestXc(t *testing.T) {
	HotReload()

	for {
		Info("abc")
		Debug(123)
		time.Sleep(2 * time.Second)
	}
}
