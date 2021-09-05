package files

import (
	"bufio"
	"encoding/json"
	"os"

	"github.com/p7chkn/go-musthave-shortener-tpl/cmd/shortener/configuration"
)

type Row struct {
	ShortURL string `json:"short_url"`
	LongURL  string `json:"long_url"`
}

func NewFileWriter() WriterInterface {
	writer, _ := newWriter(configuration.Configuration.FilePath)
	return WriterInterface(writer)
}

func NewFileReader() ReaderInterface {
	reader, _ := newReader(configuration.Configuration.FilePath)
	return ReaderInterface(reader)
}

type WriterInterface interface {
	WriteRow(event *Row) error
	Close() error
}

type ReaderInterface interface {
	ReadRow() (*Row, error)
	Close() error
}

type writer struct {
	file   *os.File
	writer *bufio.Writer
}

type reader struct {
	file    *os.File
	scanner *bufio.Scanner
}

func newWriter(fileName string) (*writer, error) {
	file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0777)
	if err != nil {
		return nil, err
	}

	return &writer{
		file:   file,
		writer: bufio.NewWriter(file),
	}, nil
}

func newReader(fileName string) (*reader, error) {
	file, err := os.OpenFile(fileName, os.O_RDONLY|os.O_CREATE, 0777)
	if err != nil {
		return nil, err
	}

	return &reader{
		file:    file,
		scanner: bufio.NewScanner(file),
	}, nil
}

func (p *writer) WriteRow(event *Row) error {
	data, err := json.Marshal(&event)
	if err != nil {
		return err
	}

	if _, err := p.writer.Write(data); err != nil {
		return err
	}

	if err := p.writer.WriteByte('\n'); err != nil {
		return err
	}

	return p.writer.Flush()
}

func (c *reader) ReadRow() (*Row, error) {
	if !c.scanner.Scan() {
		return nil, c.scanner.Err()
	}
	data := c.scanner.Bytes()

	event := &Row{}
	err := json.Unmarshal(data, event)
	if err != nil {
		return nil, err
	}

	return event, nil
}

func (p *writer) Close() error {
	return p.file.Close()
}

func (c *reader) Close() error {
	return c.file.Close()
}
