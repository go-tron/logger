package logger

import (
	"testing"
)

func TestNew(t *testing.T) {
	s := NewZap("nsq-consumer", "error")
	s.Info("hi")
}
