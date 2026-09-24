package output

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/hlebecz/crawler-cli/internal/model"
)

type fileOutput struct {
	File *os.File
}

func NewFileOutput(f *os.File) fileOutput {
	return fileOutput{File: f}
}

func (f fileOutput) Output(n model.Node) error {
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

func (f fileOutput) Close() error {
	return f.File.Close()
}
