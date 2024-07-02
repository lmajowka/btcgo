package utils

// Quantos CPUs gostaria de usar?
func ReadCPUsForUse() int {
	requestStr := "\n\nQuantos CPUs gostaria de usar?: "
	errorStr := "Numero invalido."
	return PromptForIntInRange(requestStr, errorStr, 1, 50)
}
