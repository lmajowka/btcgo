package utils

import (
    "bufio"
    "fmt"
    "os"
    "runtime"
    "strconv"
    "strings"
)

func AutoPrompt(asking string, totalNumbers int) int {
    reader := bufio.NewReader(os.Stdin)
    charReadline := '\n'

    if runtime.GOOS == "windows" {
        charReadline = '\r'
    }

    for {
        fmt.Printf(asking)
        input, _ := reader.ReadString(byte(charReadline))
        input = strings.TrimSpace(input)
        answer, err := strconv.Atoi(input)
        if err == nil && answer >= 1 && answer <= totalNumbers {
            return answer
        }
        fmt.Println("Resposta inválida.")
    }
}
