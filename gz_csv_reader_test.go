package main

import (
	"github.com/stretchr/testify/assert"
	"io"
	"os"
	"testing"
)

func TestGzCsvReader(t *testing.T) {
	dir, _ := os.Getwd()
	dataFileName := dir + "/gz_csv_reader_test_data.csv.gz"
	reader, err := NewGzCsvReader(dataFileName)
	assert.NoError(t, err)
	l1, err := reader.ReadLine()
	assert.NoError(t, err)
	assert.Equal(t, "id1", l1["id"])
	assert.Equal(t, "1970-10-26 08:48:42", l1["last_active"])
	l2, err := reader.ReadLine()
	assert.NoError(t, err)
	assert.Equal(t, "id2", l2["id"])
	assert.Equal(t, "2099-10-26 08:48:42", l2["last_active"])
	l3, err := reader.ReadLine()
	assert.ErrorIs(t, err, io.EOF)
	assert.Nil(t, l3)
}
