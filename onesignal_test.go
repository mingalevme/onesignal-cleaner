package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestArrayName(t *testing.T) {
	arrayEmailSender := NewArrayEmailSender()
	assert.Equal(t, "array", arrayEmailSender.Name())
}
