package main

import (
	"bufio"
	"fmt"
	"math/big"
	"math/rand/v2"
	"os"
	"strconv"
	"strings"
)

// promptRangeNumber prompts the user to select a range number
func PromptRangeNumber(totalRanges int) int {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Printf("Escolha a carteira (1 a %d): ", totalRanges)
		input, _ := reader.ReadString(byte(CharReadline))
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
	for {
		fmt.Printf("Quantos CPUs gostaria de usar : ")
		input, _ := reader.ReadString(byte(CharReadline))
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
	for {
		fmt.Printf("1 - Modo do inicio "+CharNewLine+"2 - Modo sequencial (chave do arquivo)"+CharNewLine+"3 - Random Steps"+CharNewLine+"4 - Random"+CharNewLine+"Escolha os modos que deseja de (1 a %d) : ", totalModos)
		input, _ := reader.ReadString(byte(CharReadline))
		input = strings.TrimSpace(input)
		modoSelecinado, err := strconv.Atoi(input)
		if err == nil && modoSelecinado >= 1 && modoSelecinado <= totalModos {
			return modoSelecinado
		}
		fmt.Println("Modo invalido.")
	}
}

func PromptMaxStep() int {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("Valor Maximo Random : ")
		input, _ := reader.ReadString(byte(CharReadline))
		input = strings.TrimSpace(input)
		valor, _ := strconv.Atoi(input)
		return valor
	}
}

// PromptAuto solicita ao usuário a seleção de um número dentro de um intervalo específico.
func PromptAuto(pergunta string, totalnumbers int) int {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print(pergunta)
		input, _ := reader.ReadString(byte(CharReadline))
		input = strings.TrimSpace(input)
		resposta, err := strconv.Atoi(input)
		if err == nil && resposta >= 1 && resposta <= totalnumbers {
			return resposta
		}
		fmt.Println("Resposta inválida.")
	}
}

// HandleModoSelecionado - selecionar modos de incializacao
func HandleModoSelecionado(modoSelecionado int, ranges *Ranges, rangeNumber int, carteirasalva string) *big.Int {
	privKeyInt := new(big.Int)
	switch modoSelecionado {
	case 1:
		// Initialize privKeyInt with the minimum value of the selected range
		privKeyHex := ranges.Ranges[rangeNumber-1].Min
		privKeyInt.SetString(privKeyHex[2:], 16)
	case 2:
		privKeyMin := ranges.Ranges[rangeNumber-1].Min
		privKeyMax := ranges.Ranges[rangeNumber-1].Max
		PrivKeyMinInt.SetString(privKeyMin[2:], 16)
		PrivKeyMaxInt.SetString(privKeyMax[2:], 16)

		verificaKey, err := LoadUltimaKeyWallet("ultimaChavePorCarteira.txt", carteirasalva)
		if err != nil || verificaKey == "" {
			// FAZER PERGUNTA SE DESEJA INFORMAR O NUMERO DE INCIO DO MODO SEQUENCIAL OU COMEÇAR DO INICIO
			msSequencialouInicio := PromptAuto("Opção 1: Deseja começar do inicio da busca (não efetivo)"+CharNewLine+"Opção 2: Escolher entre o range da carteira informada"+CharNewLine+"Por favor numero entre 1 ou 2 : ", 2)
			if msSequencialouInicio == 2 {
				// Definindo as variáveis privKeyMinInt e privKeyMaxInt como big.Int
				privKeyMin := ranges.Ranges[rangeNumber-1].Min
				privKeyMax := ranges.Ranges[rangeNumber-1].Max
				PrivKeyMinInt.SetString(privKeyMin[2:], 16)
				PrivKeyMaxInt.SetString(privKeyMax[2:], 16)

				// Calculando a diferença entre privKeyMaxInt e privKeyMinInt
				RangeKey = new(big.Int).Sub(PrivKeyMaxInt, PrivKeyMinInt)
				// Solicitando a porcentagem do range da carteira como entrada
				var rangeCarteiraSequencialStr string
				fmt.Print("Informe a porcentagem do range da carteira entre 1 a 100 : ")
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
				verificaKey = CalcPrivKey(big.NewFloat(rangeCarteiraSequencial / 100.0))
				privKeyInt.SetString(verificaKey, 16)
				fmt.Printf("Range informado, iniciando : %s"+CharNewLine, verificaKey)
			} else {
				verificaKey = ranges.Ranges[rangeNumber-1].Min
				privKeyInt.SetString(verificaKey[2:], 16)
				fmt.Printf("Nenhuma chave privada salva encontrada, iniciando do começo. %s:"+CharNewLine+" %s"+CharNewLine, carteirasalva, verificaKey)
			}
		} else {
			fmt.Printf("Encontrada chave no arquivo ultimaChavePorCarteira.txt pela carteira %s:"+CharNewLine+" %s"+CharNewLine, carteirasalva, verificaKey)
			privKeyInt.SetString(verificaKey, 16)
			rangeDiff := new(big.Int)
			walletDiff := new(big.Int)
			rangeDiff.Sub(PrivKeyMaxInt, PrivKeyMinInt)
			walletDiff.Sub(privKeyInt, PrivKeyMinInt)
			percentage := new(big.Float).Quo(new(big.Float).SetInt(walletDiff), new(big.Float).SetInt(rangeDiff))
			percentage.Mul(percentage, big.NewFloat(100))
			porcentRange := new(big.Float)
			porcentRange.SetString(percentage.String())
			fmt.Printf("A porcentagem dentro do range está em %.2f%%."+CharNewLine, porcentRange)
		}
	case 3:
		StepValue = PromptMaxStep()
		if StepValue == 0 {
			StepValue = 1
		}
		verificaKey, err := LoadUltimaKeyWallet("ultimaChavePorCarteira.txt", carteirasalva)
		if err != nil || verificaKey == "" {
			privKeyMin := ranges.Ranges[rangeNumber-1].Min
			privKeyMax := ranges.Ranges[rangeNumber-1].Max
			PrivKeyMinInt.SetString(privKeyMin[2:], 16)
			PrivKeyMaxInt.SetString(privKeyMax[2:], 16)
			// Calculando a diferença entre privKeyMaxInt e privKeyMinInt
			RangeKey = new(big.Int).Sub(PrivKeyMaxInt, PrivKeyMinInt)
			verificaKey = CalcPrivKey(big.NewFloat(float64(50) / 100))
			privKeyInt.SetString(verificaKey, 16)
			fmt.Printf("Range informado, iniciando : %s"+CharNewLine, verificaKey)
		} else {
			fmt.Printf("Encontrada chave no arquivo ultimaChavePorCarteira.txt pela carteira %s:"+CharNewLine+" %s"+CharNewLine, carteirasalva, verificaKey)
			privKeyInt.SetString(verificaKey, 16)
		}
	case 4:
		privKeyMin := ranges.Ranges[rangeNumber-1].Min
		privKeyMax := ranges.Ranges[rangeNumber-1].Max
		PrivKeyMinInt.SetString(privKeyMin[2:], 16)
		PrivKeyMaxInt.SetString(privKeyMax[2:], 16)
		// Calculando a diferença entre privKeyMaxInt e privKeyMinInt
		RangeKey = new(big.Int).Sub(PrivKeyMaxInt, PrivKeyMinInt)
		verificaKey := CalcPrivKey(big.NewFloat(randFloat(1, 99) / 100))
		privKeyInt.SetString(verificaKey, 16)
		fmt.Printf("Range informado, iniciando : %s"+CharNewLine, verificaKey)
	}
	return privKeyInt
}

func randFloat(min, max float64) float64 {
	return (min + rand.Float64()*(max-min))
}

// Calcula a PrivKey
func CalcPrivKey(rangeCarteiraSequencial *big.Float) string {
	rangeMultiplier := new(big.Float).Mul(new(big.Float).SetInt(RangeKey), rangeCarteiraSequencial)
	// Convertendo o resultado para inteiro (arredondamento para baixo)
	min := new(big.Int)
	rangeMultiplier.Int(min)
	// Adicionando rangeMultiplier ao valor mínimo (privKeyMinInt)
	min.Add(PrivKeyMinInt, min)
	// Verificando o valor final como uma string hexadecimal
	verificaKey := min.Text(16)
	return verificaKey
}
