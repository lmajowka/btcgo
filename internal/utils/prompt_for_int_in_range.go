package utils

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
)

// getInputReader retorna um bufio.Reader para leitura da entrada padrão e o caractere de nova linha
// apropriado, que é '\r' no Windows e '\n' em outros sistemas operacionais.
func getInputReader() (*bufio.Reader, rune) {
	reader := bufio.NewReader(os.Stdin)
	charReadline := '\n'

	if runtime.GOOS == "windows" {
		charReadline = '\r'
	}
	return reader, charReadline
}

// PromptAuto solicita ao usuário a seleção de um número dentro de um intervalo específico.
// Sera retornado um numero que necessariamente atenda a (min <= X <= max) onde X foi o numero escolhido do usuario
func PromptForIntInRange(request_str string, error_str string, mim int, max int) int {
	reader, charReadline := getInputReader()

	for {
		fmt.Printf(request_str)
		input, _ := reader.ReadString(byte(charReadline))
		input = strings.TrimSpace(input)
		resposta, err := strconv.Atoi(input)
		if err == nil && resposta >= mim && resposta <= max {
			return resposta
		}
		fmt.Println(error_str)
	}
}
