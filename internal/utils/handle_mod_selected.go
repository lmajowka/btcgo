package utils

import (
	"btcgo/internal/domain"
	"fmt"
	"math/big"
	"strconv"
	"strings"
)

func HandleModSelected(modoSelecionado int, ranges *domain.Ranges, rangeNumber int, privKeyInt *big.Int, carteirasalva string) *big.Int {
	if modoSelecionado == 1 {
		// Initialize privKeyInt with the minimum value of the selected range
		privKeyHex := ranges.Ranges[rangeNumber-1].Min
		privKeyInt.SetString(privKeyHex[2:], 16)
	} else if modoSelecionado == 2 {
		verificaKey, err := LoadLastKeyWallet("ultimaChavePorCarteira.txt", carteirasalva)
		if err != nil || verificaKey == "" {
			// FAZER PERGUNTA SE DESEJA INFORMAR O NUMERO DE INCIO DO MODO SEQUENCIAL OU COMEÇAR DO INICIO
			msSequencialouInicio := PromptForIntInRange(
				"Opção 1: Deseja começar do inicio da busca (não efetivo) ou \nOpção 2: Escolher entre o range da carteira informada? \nPor favor numero entre 1 ou 2: ",
				"Número inválido. Escolha entre 1 ou 2.",
				1, 2)
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
