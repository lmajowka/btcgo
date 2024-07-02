package utils

import (
    "btcgo/internal/domain"
    "encoding/json"
    "io"
    "os"
)

func LoadRanges(filename string) (*domain.Ranges, error) {
    file, err := os.Open(filename)
    if err != nil {
        return nil, err
    }
    defer file.Close()

    bytes, err := io.ReadAll(file)
    if err != nil {
        return nil, err
    }

    var ranges domain.Ranges
    if err := json.Unmarshal(bytes, &ranges); err != nil {
        return nil, err
    }

    return &ranges, nil
}
