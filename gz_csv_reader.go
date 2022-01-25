package main

import (
	"compress/gzip"
	"encoding/csv"
	"github.com/pkg/errors"
	"io"
	"os"
)

type GzCsvReader struct {
	Filename  string
	file      *os.File
	reader    *gzip.Reader
	csvReader *csv.Reader
	header    []string
}

func NewGzCsvReader(filename string) (*GzCsvReader, error) {
	reader := &GzCsvReader{Filename: filename}
	if err := reader.init(); err != nil {
		reader.Close()
		return nil, err
	}
	return reader, nil
}

func (r *GzCsvReader) init() error {
	f, err := os.Open(r.Filename)
	if err != nil {
		return errors.Wrap(err, "error while opening players data file")
	}
	r.file = f
	gr, err := gzip.NewReader(f)
	if err != nil {
		return errors.Wrap(err, "error while creating a new gzip reader")
	}
	r.reader = gr
	r.csvReader = csv.NewReader(gr)
	header, err := r.csvReader.Read()
	if err != nil {
		return errors.Wrap(err, "error while reading header line")
	}
	r.header = header
	return nil
}

func (r *GzCsvReader) ReadLine() (map[string]string, error) {
	rec, err := r.csvReader.Read()
	if err != nil {
		if err == io.EOF {
			return nil, err
		}
		return nil, errors.Wrap(err, "error while reading line")
	}
	data := map[string]string{}
	for i, v := range rec {
		k := r.header[i]
		data[k] = v
	}
	return data, nil
}

func (r *GzCsvReader) Close() {
	if r.file != nil {
		_ = r.file.Close()
	}
	if r.reader != nil {
		_ = r.reader.Close()
	}
}
