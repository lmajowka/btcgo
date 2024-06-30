package utils

import (
    "fmt"
    "os"
    "strings"
)

func SaveLastKeyWallet(filename string, wallet string, key string) error {
    // abre o arquivo em modo de append, cria se não existir
    file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0644)
    if err != nil {
        return err
    }
    defer file.Close()

    // busca todo o conteúdo atual do arquivo
    data, err := os.ReadFile(filename)
    if err != nil {
        return err
    }

    lines := strings.Split(string(data), "\n")
    found := false

    // verifica se a carteira já existe no arquivo
    for i, line := range lines {
        parts := strings.SplitN(line, "|", 2) // divide a linha em duas partes pelo " | "
        if len(parts) != 2 {
            continue // ignora linhas mal formatadas sem " | "
        }
        if parts[0] == wallet {
            // Substitui apenas a chave correspondente à carteira encontrada
            lines[i] = fmt.Sprintf("%s|%s", wallet, key)
            found = true
            break
        }
    }
    // se a carteira não foi encontrada, adiciona uma nova linha
    if !found {
        lines = append(lines, fmt.Sprintf("%s|%s", wallet, key))
    }
    // salva no arquivo com as modificações
    err = os.WriteFile(filename, []byte(strings.Join(lines, "\n")), 0644)
    if err != nil {
        return err
    }

    return nil
}
