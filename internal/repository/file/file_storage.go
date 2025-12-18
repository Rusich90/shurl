package file

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"

	"github.com/Rusich90/shurl.git/internal/transport/http/dto"
)

type FileStorage struct {
	fileName string
}

func NewFileStorage(fileName string) (*FileStorage, error) {
	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, fmt.Errorf("failed os.OpenFile: %w", err)
	}
	file.Close()

	return &FileStorage{
		fileName: fileName,
	}, nil
}

func (f *FileStorage) SaveRow(row dto.URLRow) error {
	file, err := os.OpenFile(f.fileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return fmt.Errorf("failed os.OpenFile: %w", err)
	}
	defer file.Close()

	data, err := json.Marshal(row)
	if err != nil {
		return fmt.Errorf("failed json.Marshal: %w", err)
	}

	_, err = file.Write(append(data, '\n'))
	if err != nil {
		return fmt.Errorf("failed file.Write: %w", err)
	}
	return nil
}

func (f *FileStorage) GetURLs() ([]dto.URLRow, error) {
	file, err := os.Open(f.fileName)
	if err != nil {
		if os.IsNotExist(err) {
			return []dto.URLRow{}, nil
		}
		return nil, fmt.Errorf("failed os.Open: %w", err)
	}
	defer file.Close()

	var urls []dto.URLRow
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		var row dto.URLRow
		line := scanner.Text()
		if line != "" {
			if err := json.Unmarshal([]byte(line), &row); err != nil {
				continue
			}
			urls = append(urls, row)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scanner.Err: %w", err)
	}

	return urls, nil
}
