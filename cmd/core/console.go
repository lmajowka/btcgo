package core

import (
	"btcgo/cmd/utils"
	"bufio"
	"fmt"
	"github.com/fatih/color"
	"github.com/joho/godotenv"
	"os"
	"runtime"
	"strconv"
	"strings"
)

const (
	MaxModos = 3
)

func RequestData() {
	err := godotenv.Load(".env")

	if err != nil {
		handleEnvLoadError()
	} else {
		initializeFromEnv()
	}

	handleModo()
}

func handleEnvLoadError() {
	green := color.New(color.FgGreen).SprintFunc()
	fmt.Printf("\nCPUs detectados: %s", green(runtime.NumCPU()))
	App.MaxWorkers = readCPUsForUse()
	App.RangeNumber = promptRangeNumber()
	App.Carteira = fmt.Sprintf("%d", App.RangeNumber)
	App.Wallets.SetFindWallet(App.RangeNumber)
	App.Modo = promptMods(MaxModos)
}

func initializeFromEnv() {
	numCPU := getEnvInt("CPU", runtime.NumCPU())
	wallet := getEnvInt("WALLET", -1)
	mode := getEnvInt("MODE", -1)

	if numCPU <= runtime.NumCPU() {
		App.MaxWorkers = numCPU
	} else {
		fmt.Print("O número informado de processadores excede ao existente no dispositivo.")
	}

	App.RangeNumber = wallet
	if mode > MaxModos {
		fmt.Println("Modo inválido. Escolha entre 1, 2 ou 3.")
	}
	App.Modo = mode
	switch App.Modo {
	case 2:
		App.DesdeInicio = true
		handleSequentialMode()
	case 2:
		App.DesdeInicio = false
		handleSequentialMode()
	case 3:
		handleRandomMode()
	}
}

func handleSequentialMode() {
	msSequencialouInicio := promptForIntInRange(
		"\n\nOpção 1: Deseja começar do inicio da busca (não efetivo) ou \nOpção 2: Escolher entre o range da carteira informada? \n\nPor favor numero entre 1 ou 2: ",
		"Número inválido. Escolha entre 1 ou 2.",
		1, 2)
	if msSequencialouInicio == 1 {
		App.DesdeInicio = true
	} else {
		_, err := App.LastKey.GetLastKey(App.Carteira)
		if err != nil {
			rangeCarteiraSequencialStr := promptPercentage()
			App.StartPosPercent, _ = strconv.ParseFloat(rangeCarteiraSequencialStr, 64)
		}
	}
}

func handleRandomMode() {
	App.DesdeInicio = false
	db := getEnvInt("DB", -1)
	registry := getEnvInt("RANDON_REGISTRY", -1)
	App.USEDB = db
	App.Keys.SetRecs(registry)
}

func getEnvInt(envVar string, defaultValue int) int {
	valStr := os.Getenv(envVar)
	val, err := strconv.Atoi(valStr)
	if err != nil {
		fmt.Printf("Valor inválido para %s: %v\n", envVar, err)
		return defaultValue
	}
	return val
}

func promptPercentage() string {
	var rangeCarteiraSequencialStr string
	fmt.Print("Informe a porcentagem do range da carteira entre 1 a 100: ")
	fmt.Scanln(&rangeCarteiraSequencialStr)
	return strings.Replace(rangeCarteiraSequencialStr, ",", ".", -1)
}

func readCPUsForUse() int {
	requestStr := "\n\nQuantos CPUs gostaria de usar?: "
	errorStr := "Numero invalido."
	return promptForIntInRange(requestStr, errorStr, 1, 50)
}

func promptRangeNumber() int {
	totalRanges := App.Ranges.Count()
	requestStr := fmt.Sprintf("\n\nEscolha a carteira (1 a %d): ", totalRanges)
	errorStr := "Numero invalido."
	return promptForIntInRange(requestStr, errorStr, 1, totalRanges)
}

func promptMods(totalModos int) int {
	requestStr := fmt.Sprintf("\n\nEscolha os modos que deseja de (1 a %d)\n\nModo do inicio: 1\nModo sequencial(chave do arquivo): 2\nModo Random: 3\n\nEscolha o modo: ", totalModos)
	errorStr := "Modo invalido."
	return promptForIntInRange(requestStr, errorStr, 1, totalModos)
}

func promptUseDB(totalModos int) int {
	requestStr := "\nUtiliza BaseDados para controlar repetiçóes?\nModo Random com DB: 1\nModo Random sem DB: 2\n\nEscolha o modo: "
	errorStr := "Modo invalido."
	return promptForIntInRange(requestStr, errorStr, 1, totalModos)
}

func promptNumRecsRandom() int {
	requestStr := "\nNumero registos por cada random (ex. 10000): "
	errorStr := "Modo invalido."
	return promptForIntInRange(requestStr, errorStr, 1, 0)
}

func promptForIntInRange(requestStr string, errorStr string, min int, max int) int {
	charReadline := utils.GetEndLineChar()
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print(requestStr)
		input, _ := reader.ReadString(byte(charReadline))
		input = strings.TrimSpace(input)
		resposta, err := strconv.Atoi(input)
		if max == 0 {
			if err == nil && resposta >= min {
				return resposta
			}
		}
		if err == nil && resposta >= min && resposta <= max {
			return resposta
		}
		fmt.Println(errorStr)
	}
}
