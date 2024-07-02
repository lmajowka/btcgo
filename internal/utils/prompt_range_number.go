package utils

import (
	"fmt"
)

// promptRangeNumber prompts the user to select a range number
func PromptRangeNumber(totalRanges int) int {
	requestStr := fmt.Sprintf("\n\nEscolha a carteira (1 a %d): ", totalRanges)
	errorStr := "Numero invalido."
	return PromptForIntInRange(requestStr, errorStr, 1, totalRanges)
}

// PromptModos prompts the user to select a modo's
func PromptMods(totalModos int) int {
	requestStr := fmt.Sprintf("\n\nEscolha os modos que deseja de (1 a %d)\n\nModo do inicio: 1\nModo sequencial(chave do arquivo): 2\n\nEscolha o modo: ", totalModos)
	errorStr := "Modo invalido."
	return PromptForIntInRange(requestStr, errorStr, 1, totalModos)
}
