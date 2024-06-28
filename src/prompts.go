package main

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
)

// promptRangeNumber prompts the user to select a range number
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

func PromptCPUNumber() int {
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


// PromptModos prompts the user to select a modo's
func PromptModos(totalModos int) int {
	reader := bufio.NewReader(os.Stdin)
	charReadline := '\n'

	if runtime.GOOS == "windows" {
		charReadline = '\r'
	}

	for {
		fmt.Printf("Escolha os modos que deseja de (1 a %d) \n  Modo do inicio: 1 - Modo sequencial(chave do arquivo): 2): ", totalModos)
		input, _ := reader.ReadString(byte(charReadline))
		input = strings.TrimSpace(input)
		modoSelecinado, err := strconv.Atoi(input)
		if err == nil && modoSelecinado >= 1 && modoSelecinado <= totalModos {
			return modoSelecinado
			//fmt.Println(modoSelecinado)
		}
		fmt.Println("Modo invalido.")
	}
}

// PromptAuto solicita ao usuário a seleção de um número dentro de um intervalo específico.
func PromptAuto(pergunta string, totalnumbers int) int {
	reader := bufio.NewReader(os.Stdin)
	charReadline := '\n'

	if runtime.GOOS == "windows" {
		charReadline = '\r'
	}

	for {
		fmt.Printf(pergunta)
		input, _ := reader.ReadString(byte(charReadline))
		input = strings.TrimSpace(input)
		resposta, err := strconv.Atoi(input)
		if err == nil && resposta >= 1 && resposta <= totalnumbers {
			return resposta
		}
		fmt.Println("Resposta inválida.")
	}
}