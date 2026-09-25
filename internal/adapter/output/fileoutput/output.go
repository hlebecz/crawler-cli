package fileoutput

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/hlebecz/crawler-cli/internal/model"
)

type FileOutput struct {
	File *os.File
}

func New(f *os.File) FileOutput {
	return FileOutput{File: f}
}

func (f FileOutput) Output(n []*model.Node) error {
	body, err := json.Marshal(n)
	if err != nil {
		return fmt.Errorf("unable to marshal node: %w", err)
	}

	_, err = f.File.Write(body)
	if err != nil {
		return fmt.Errorf("unable to write node: %w", err)
	}
	return nil
}

func (f FileOutput) Close() error {
	return f.File.Close()
}
