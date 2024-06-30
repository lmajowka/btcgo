package main

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"math/big"
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

//perguntar se deseja verificar todas as carteiras ou apenas uma
func PromptUniqueOrAll() int {
	reader := bufio.NewReader(os.Stdin)
	charReadline := '\n'

	if runtime.GOOS == "windows" {
		charReadline = '\r'
	}

	for {
		fmt.Printf("Escolha uma opção:\n 1. Verificar todas as carteiras a cada geração (mais chance) \n 2. Verificar uma carteira por vez (mais desempenho) \n> ")
		input, _ := reader.ReadString(byte(charReadline))
		input = strings.TrimSpace(input)
		modoSelecinado, err := strconv.Atoi(input)
		if err == nil && (modoSelecinado == 1 || modoSelecinado == 2) {
			return modoSelecinado
		}
		fmt.Println("Modo invalido.")
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

// HandleModoSelecionado - selecionar modos de incializacao
func HandleModoSelecionado(modoSelecionado int, ranges *Ranges, rangeNumber int, privKeyInt *big.Int, carteirasalva string) *big.Int {
    if modoSelecionado == 1 {
        // Initialize privKeyInt with the minimum value of the selected range
        privKeyHex := ranges.Ranges[rangeNumber-1].Min
        privKeyInt.SetString(privKeyHex[2:], 16)

    } else if modoSelecionado == 2 {
        verificaKey, err := LoadUltimaKeyWallet("ultimaChavePorCarteira.txt", carteirasalva)
        if err != nil || verificaKey == "" {
            // FAZER PERGUNTA SE DESEJA INFORMAR O NUMERO DE INCIO DO MODO SEQUENCIAL OU COMEÇAR DO INICIO
            msSequencialouInicio := PromptAuto("Opção 1: Deseja começar do inicio da busca (não efetivo) ou \nOpção 2: Escolher entre o range da carteira informada? \nPor favor numero entre 1 ou 2:", 2)
            if msSequencialouInicio == 2 {
                // Definindo as variáveis privKeyMinInt e privKeyMaxInt como big.Int
                privKeyMinInt := new(big.Int)
                privKeyMaxInt := new(big.Int)
                privKeyMin := ranges.Ranges[rangeNumber-1].Min
                privKeyMax := ranges.Ranges[rangeNumber-1].Max
                privKeyMinInt.SetString(privKeyMin[2:], 16)
                privKeyMaxInt.SetString(privKeyMax[2:], 16)

                // Calculando a diferença entre privKeyMaxInt e privKeyMinInt
                rangeKey := new(big.Int).Sub(privKeyMaxInt, privKeyMinInt)

                // Solicitando a porcentagem do range da carteira como entrada
                var rangeCarteiraSequencialStr string
                fmt.Print("Informe a porcentagem do range da carteira entre 1 a 100: ")
                fmt.Scanln(&rangeCarteiraSequencialStr)

                // Substituindo vírgulas por pontos se necessário
                rangeCarteiraSequencialStr = strings.Replace(rangeCarteiraSequencialStr, ",", ".", -1)

                // Convertendo a porcentagem para um número decimal
                rangeCarteiraSequencial, err := strconv.ParseFloat(rangeCarteiraSequencialStr, 64)
                if err != nil {
                    fmt.Println("Erro ao ler porcentagem:", err)
                    return nil
                }

                // Verificando se a porcentagem está no intervalo válido
                if rangeCarteiraSequencial < 1 || rangeCarteiraSequencial > 100 {
                    fmt.Println("Porcentagem fora do intervalo válido (1 a 100).")
                    return nil
                }

                // Calculando o valor de rangeKey multiplicado pela porcentagem
                rangeMultiplier := new(big.Float).Mul(new(big.Float).SetInt(rangeKey), big.NewFloat(rangeCarteiraSequencial/100.0))

                // Convertendo o resultado para inteiro (arredondamento para baixo)
                min := new(big.Int)
                rangeMultiplier.Int(min)

                // Adicionando rangeMultiplier ao valor mínimo (privKeyMinInt)
                min.Add(privKeyMinInt, min)

                // Verificando o valor final como uma string hexadecimal
                verificaKey := min.Text(16)
                privKeyInt.SetString(verificaKey, 16)
                fmt.Printf("Range informado, iniciando: %s\n", verificaKey)
            } else {
                verificaKey = ranges.Ranges[rangeNumber-1].Min
                privKeyInt.SetString(verificaKey[2:], 16)
                fmt.Printf("Nenhuma chave privada salva encontrada, iniciando do começo. %s: %s\n", carteirasalva, verificaKey)
            }
        } else {
            fmt.Printf("Encontrada chave no arquivo ultimaChavePorCarteira.txt pela carteira %s: %s\n", carteirasalva, verificaKey)
            privKeyInt.SetString(verificaKey, 16)
        }
    }
    return privKeyInt
}