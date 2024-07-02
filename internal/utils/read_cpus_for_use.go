package utils

import (
    "bufio"
    "fmt"
    "os"
    "runtime"
    "strconv"
    "strings"
)

func ReadCPUsForUse() int {
    reader := bufio.NewReader(os.Stdin)
    charReadline := '\n'

    if runtime.GOOS == "windows" {
        charReadline = '\r'
    }

    for {
        fmt.Printf("Quantos CPUs gostaria de usar?: ")
        input, _ := reader.ReadString(byte(charReadline))
        input = strings.TrimSpace(input)
        cpusNumber, err := strconv.Atoi(input)
        if err == nil && cpusNumber >= 1 && cpusNumber <= 50 {
            return cpusNumber
        }
        fmt.Println("Numero invalido.")
    }
}
