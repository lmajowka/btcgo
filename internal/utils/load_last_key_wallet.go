package utils

import (
    "fmt"
    "os"
    "strings"
)

func LoadLastKeyWallet(filename string, wallet string) (string, error) {
    // Busca todo o conteúdo atual do arquivo
    data, err := os.ReadFile(filename)
    if err != nil {
        return "", err
    }

    lines := strings.Split(string(data), "\n")

    // Verifica se a carteira já existe no arquivo
    for _, line := range lines {
        parts := strings.SplitN(line, "|", 2) // Divide a linha em duas partes pelo "|"
        if len(parts) != 2 {
            continue // Ignora linhas mal formatadas sem "|"
        }
        if parts[0] == wallet {
            return parts[1], nil
        }
    }

    // Retorna erro se a carteira não for encontrada
    return "", fmt.Errorf("carteira %s não encontrada", wallet)
}
