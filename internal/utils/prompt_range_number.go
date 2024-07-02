package utils

import (
    "bufio"
    "fmt"
    "os"
    "runtime"
    "strconv"
    "strings"
)

func PromptRangeNumber(totalRanges int) int {
    reader := bufio.NewReader(os.Stdin)
    charReadline := '\n'

    if runtime.GOOS == "windows" {
        charReadline = '\r'
    }

    for {
        fmt.Printf("Escolha a carteira (1 a %d): ", totalRanges)
        input, _ := reader.ReadString(byte(charReadline))
        input = strings.TrimSpace(input)
        rangeNumber, err := strconv.Atoi(input)
        if err == nil && rangeNumber >= 1 && rangeNumber <= totalRanges {
            return rangeNumber
        }
        fmt.Println("Numero invalido.")
    }
}

func PromptMods(totalMods int) int {
    reader := bufio.NewReader(os.Stdin)
    charReadline := '\n'

    if runtime.GOOS == "windows" {
        charReadline = '\r'
    }

    for {
        fmt.Printf("Escolha os modos que deseja de (1 a %d) \n  Modo do inicio: 1 - Modo sequencial(chave do arquivo): 2): ", totalMods)
        input, _ := reader.ReadString(byte(charReadline))
        input = strings.TrimSpace(input)
        modSelected, err := strconv.Atoi(input)
        if err == nil && modSelected >= 1 && modSelected <= totalMods {
            return modSelected
        }
        fmt.Println("Modo invalido.")
    }
}
