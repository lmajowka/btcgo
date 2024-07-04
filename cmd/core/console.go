/**
 * BTCGO
 *
 * Modulo : Consola
 */

package core

import (
	"btcgo/cmd/utils"
	"bufio"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/fatih/color"
)

func RequestData() {
	green := color.New(color.FgGreen).SprintFunc()
	// Number of CPU cores to use
	fmt.Printf("\nCPUs detectados: %s", green(runtime.NumCPU()))
	App.MaxWorkers = readCPUsForUse()

	// Ask the user for the range number
	App.RangeNumber = promptRangeNumber()
	App.Carteira = fmt.Sprintf("%d", App.RangeNumber)

	// Set Search Wallet address
	App.Wallets.SetFindWallet(App.RangeNumber)

	// Pergunta sobre modos de usar
	App.Modo = promptMods(2) // quantidade de modos

	if App.Modo == 2 {
		App.DesdeInicio = false
		msSequencialouInicio := promptForIntInRange(
			"\n\nOpção 1: Deseja começar do inicio da busca (não efetivo) ou \nOpção 2: Escolher entre o range da carteira informada? \n\nPor favor numero entre 1 ou 2: ",
			"Número inválido. Escolha entre 1 ou 2.",
			1, 2)
		if msSequencialouInicio == 1 {
			App.DesdeInicio = true
		} else {
			_, err := App.LastKey.GetLastKey(App.Carteira)
			if err != nil {
				// Solicitando a porcentagem do range da carteira como entrada
				var rangeCarteiraSequencialStr string
				fmt.Print("Informe a porcentagem do range da carteira entre 1 a 100: ")
				fmt.Scanln(&rangeCarteiraSequencialStr)
				// Substituindo vírgulas por pontos se necessário
				rangeCarteiraSequencialStr = strings.Replace(rangeCarteiraSequencialStr, ",", ".", -1)
				App.StartPosPercent, _ = strconv.ParseFloat(rangeCarteiraSequencialStr, 64)
			}
		}
	}
}

// Quantos CPUs gostaria de usar?
func readCPUsForUse() int {
	requestStr := "\n\nQuantos CPUs gostaria de usar?: "
	errorStr := "Numero invalido."
	return promptForIntInRange(requestStr, errorStr, 1, 50)
}

// promptRangeNumber prompts the user to select a range number
func promptRangeNumber() int {
	totalRanges := App.Ranges.Count()
	requestStr := fmt.Sprintf("\n\nEscolha a carteira (1 a %d): ", totalRanges)
	errorStr := "Numero invalido."
	return promptForIntInRange(requestStr, errorStr, 1, totalRanges)
}

// PromptModos prompts the user to select a modo's
func promptMods(totalModos int) int {
	requestStr := fmt.Sprintf("\n\nEscolha os modos que deseja de (1 a %d)\n\nModo do inicio: 1\nModo sequencial(chave do arquivo): 2\n\nEscolha o modo: ", totalModos)
	errorStr := "Modo invalido."
	return promptForIntInRange(requestStr, errorStr, 1, totalModos)
}

// PromptAuto solicita ao usuário a seleção de um número dentro de um intervalo específico.
// Sera retornado um numero que necessariamente atenda a (min <= X <= max) onde X foi o numero escolhido do usuario
func promptForIntInRange(request_str string, error_str string, mim int, max int) int {
	charReadline := utils.GetEndLineChar()
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print(request_str)
		input, _ := reader.ReadString(byte(charReadline))
		input = strings.TrimSpace(input)
		resposta, err := strconv.Atoi(input)
		if err == nil && resposta >= mim && resposta <= max {
			return resposta
		}
		fmt.Println(error_str)
	}
}
