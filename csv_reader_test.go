package main

import (
	"github.com/stretchr/testify/assert"
	"io"
	"os"
	"testing"
)

func TestGzCsvReader(t *testing.T) {
	dir, _ := os.Getwd()
	dataFileName := dir + "/csv_reader_test_data.csv.gz"
	reader, err := NewGzCsvReader(dataFileName)
	assert.NoError(t, err)
	l1, err := reader.ReadLine()
	assert.NoError(t, err)
	assert.Equal(t, "value11", l1["header1"])
	assert.Equal(t, "value12", l1["header2"])
	l2, err := reader.ReadLine()
	assert.NoError(t, err)
	assert.Equal(t, "value21", l2["header1"])
	assert.Equal(t, "value22", l2["header2"])
	l3, err := reader.ReadLine()
	assert.ErrorIs(t, err, io.EOF)
	assert.Nil(t, l3)
}
