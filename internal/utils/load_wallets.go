package utils

import (
    "btcgo/internal/domain"
    "encoding/json"
    "io"
    "os"
)

func LoadWallets(filename string) (*domain.Wallets, error) {
    file, err := os.Open(filename)
    if err != nil {
        return nil, err
    }
    defer file.Close()

    bytes, err := io.ReadAll(file)
    if err != nil {
        return nil, err
    }

    type WalletsTemp struct {
        Addresses []string `json:"wallets"`
    }

    var walletsTemp WalletsTemp
    if err := json.Unmarshal(bytes, &walletsTemp); err != nil {
        return nil, err
    }

    var wallets domain.Wallets
    for _, address := range walletsTemp.Addresses {
        wallets.Addresses = append(wallets.Addresses, Decode(address)[1:21])
    }

    return &wallets, nil
}
