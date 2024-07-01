package main

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"os"
	"os/exec"
	"runtime"
)

func titulo() {
	fmt.Println("\x1b[38;2;250;128;114m" + "╔═══════════════════════════════════════╗")
	fmt.Println("║\x1b[0m\x1b[36m" + "   ____ _______ _____    _____  ____   " + "\x1b[0m\x1b[38;2;250;128;114m" + "║")
	fmt.Println("║\x1b[0m\x1b[36m" + "  |  _ \\__   __/ ____|  / ____|/ __ \\  " + "\x1b[0m\x1b[38;2;250;128;114m" + "║")
	fmt.Println("║\x1b[0m\x1b[36m" + "  | |_) | | | | |      | |  __| |  | | " + "\x1b[0m\x1b[38;2;250;128;114m" + "║")
	fmt.Println("║\x1b[0m\x1b[36m" + "  |  _ <  | | | |      | | |_ | |  | | " + "\x1b[0m\x1b[38;2;250;128;114m" + "║")
	fmt.Println("║\x1b[0m\x1b[36m" + "  | |_) | | | | |____  | |__| | |__| | " + "\x1b[0m\x1b[38;2;250;128;114m" + "║")
	fmt.Println("║\x1b[0m\x1b[36m" + "  |____/  |_|  \\_____|  \\_____|\\____/  " + "\x1b[0m\x1b[38;2;250;128;114m" + "║")
	fmt.Println("║\x1b[0m\x1b[36m" + "                                       " + "\x1b[0m\x1b[38;2;250;128;114m" + "║")
	fmt.Println("╚═══\x1b[32m" + " Investidor Internacional - v0.55" + "\x1b[0m\x1b[38;2;250;128;114m ══╝" + "\x1b[0m")
	fmt.Println("")
	fmt.Println(" \x1b[32m" + "Fork by Marton Lyra" + "\x1b[0m\x1b[38;2;250;128;114m" + "\x1b[0m")
	fmt.Println(" \x1b[32m" + "Com todas as funcionalidades da v0.5 original e um pouco mais" + "\x1b[0m\x1b[38;2;250;128;114m" + "\x1b[0m")
	fmt.Println(" \x1b[32m" + "https://github.com/MartonLyra/btcgo" + "\x1b[0m\x1b[38;2;250;128;114m" + "\x1b[0m")
	fmt.Println("")
}

func ClearConsole() {
	switch runtime.GOOS {
	case "windows":
		cmd := exec.Command("cmd", "/c", "cls")
		cmd.Stdout = os.Stdout
		cmd.Run()
	default:
		cmd := exec.Command("clear")
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
}

func updatePrivKeyInt(privKeyInt, rangeMinInt, rangeMaxInt *big.Int, modoSelecionado int) {
	rangeSize := new(big.Int).Sub(rangeMaxInt, rangeMinInt)
	randomNum, err := rand.Int(rand.Reader, rangeSize)
	if err != nil {
		fmt.Println("Erro ao gerar número aleatório:", err)
		return
	}
	privKeyInt.Add(randomNum, rangeMinInt)

	posicaoPercent := calculatePercentage(privKeyInt, rangeMinInt, rangeMaxInt) // Calcula a posição, em porcentagem
	posicaoPercentStr := fmt.Sprintf("%.12f", posicaoPercent)                   // Formata com 12 casas decimais

	// Representação gráfica da porcentagem:
	graphicRepresentation := generateGraphicRepresentation(posicaoPercent)

	fmt.Printf("\nMovendo para nova posição aleatória: %s ; Nova posição: %s  -  %s\n", privKeyInt.Text(16), posicaoPercentStr, graphicRepresentation)

	if modoSelecionado == 3 {
		if posicaoPercent < 4 {
			fmt.Println("Não queremos processar os primeiros 4% do range. Vamos re-calcular nova posição aleatória.")
			updatePrivKeyInt(privKeyInt, rangeMinInt, rangeMaxInt, modoSelecionado)
		}
		if posicaoPercent > 99 {
			fmt.Println("Não queremos processar os últimos 1% do range. Vamos re-calcular nova posição aleatória.")
			updatePrivKeyInt(privKeyInt, rangeMinInt, rangeMaxInt, modoSelecionado)
		}
	}

	fmt.Println()

}

// Aqui eu gero uma representação gráfica, em ASCII, da porcentagem:
func generateGraphicRepresentation(percentage float64) string {
	totalLength := 50
	position := int(percentage * float64(totalLength) / 100)

	if position >= totalLength {
		position = totalLength - 1
	} else if position < 0 {
		position = 0
	}

	representation := make([]rune, totalLength)
	for i := range representation {
		representation[i] = '_'
	}
	representation[position] = 'X'

	return fmt.Sprintf("[%s]", string(representation))
}

// Essa função calcula a posição privKeyInt representa entre rangeMinInt e rangeMaxInt, em porcentagem
func calculatePercentage(pos, min, max *big.Int) float64 {
	// Calcula o intervalo total
	totalRange := new(big.Int).Sub(max, min)

	// Calcula a posição relativa de privKeyInt dentro do intervalo
	relativePosition := new(big.Int).Sub(pos, min)

	// Converte totalRange e relativePosition para float64
	totalRangeFloat, _ := new(big.Float).SetInt(totalRange).Float64()
	relativePositionFloat, _ := new(big.Float).SetInt(relativePosition).Float64()

	// Calcula a porcentagem
	percentage := (relativePositionFloat / totalRangeFloat) * 100

	return percentage - 1
}

// Marton - A função recebe um tempo estimado para conclusão e retorna string com anos;
func formatDuration(seconds float64) string {
	// Converte os segundos em um valor inteiro
	totalSeconds := int64(seconds)

	// Define constantes para as durações
	const (
		secondsInMinute = 60
		secondsInHour   = 60 * secondsInMinute
		secondsInDay    = 24 * secondsInHour
		secondsInMonth  = 30 * secondsInDay
		secondsInYear   = 12 * secondsInMonth
	)

	years := totalSeconds / secondsInYear
	totalSeconds %= secondsInYear

	/*
		months := totalSeconds / secondsInMonth
		totalSeconds %= secondsInMonth

		days := totalSeconds / secondsInDay
		totalSeconds %= secondsInDay
	*/

	// Monta a string formatada
	formattedDuration := fmt.Sprintf("%d anos" /*, %d mes(es), %d dia(s)"*/, years /*, months, days*/)
	return formattedDuration
}
