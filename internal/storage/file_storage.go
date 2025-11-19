package storage

import (
	"bufio"
	"encoding/json"
	"os"

	"github.com/Rusich90/shurl.git/internal/model"
)

type FileStorage struct {
	fileName string
}

func NewFileStorage(fileName string) (*FileStorage, error) {
	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}
	file.Close()

	return &FileStorage{
		fileName: fileName,
	}, nil
}

func (f *FileStorage) SaveRow(row model.URLRow) error {
	file, err := os.OpenFile(f.fileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	data, err := json.Marshal(row)
	if err != nil {
		return err
	}

	_, err = file.Write(append(data, '\n'))
	return err
}

func (f *FileStorage) GetURLs() ([]model.URLRow, error) {
	file, err := os.Open(f.fileName)
	if err != nil {
		if os.IsNotExist(err) {
			return []model.URLRow{}, nil
		}
		return nil, err
	}
	defer file.Close()

	var urls []model.URLRow
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		var row model.URLRow
		line := scanner.Text()
		if line != "" {
			if err := json.Unmarshal([]byte(line), &row); err != nil {
				continue
			}
			urls = append(urls, row)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return urls, nil
}
