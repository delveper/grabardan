package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

func downloadFile(u string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	log.Printf("Request: %+v\n", req)

	client := new(http.Client)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	defer func() { _ = resp.Body.Close() }()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("expected OK, got status code: %d", resp.StatusCode)
	}

	return data, nil
}

func saveFile(data []byte, name string) error {
	file, err := os.Create(name)
	if err != nil {
		return err
	}

	n, err := file.ReadFrom(bytes.NewReader(data))
	if err != nil {
		return err
	}

	if n == 0 {
		return fmt.Errorf("no bytes read")
	}

	return nil
}
