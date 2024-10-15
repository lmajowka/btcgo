# BTC GO v0.7.0
[![instalação do Go no Windows](https://img.youtube.com/vi/679Zc7ZQLtI/0.jpg)](https://www.youtube.com/watch?v=679Zc7ZQLtI)

## Requisitos
  -  [Go][install-go]
  -  [Git][install-git]
  -  Terminal

# Instruções para rodar o projeto no Windows.

 * Clona o repo e brota na pasta:
```bash
git clone [git@github.com:lmajowka/btcgo.git](https://github.com/lmajowka/btcgo.git) && cd btcgo
```
 
 * Instala as parada:
```bash
go mod tidy
```

 * Faz o build do projeto:
```bash
go build -o btcgo.exe ./cmd/main.go
```

 * Executa o que foi compilado:
```bash
btcgo
```


# Instruções para rodar o projeto no Linux / MacOS.

 * Clona o repo e brota na pasta:
```bash
  git clone [git@github.com:lmajowka/btcgo.git](https://github.com/lmajowka/btcgo.git) && cd btcgo
```
 
 * Instala as parada:
```bash
 go mod tidy
```

 * Faz o build do projeto:
```bash
go build -o btcgo ./cmd/main.go
```

 * Executa o que foi compilado:
```bash
./btcgo
```

# Instruções para rodar o projeto em container.

## Requisitos
  -  [Docker][install-docker]
  -  [Docker-compose][install-docker-compose]
  -  [Git][install-git]

## Execução da parada

 * Clona o repo:
```bash
git clone git@github.com:lmajowka/btcgo.git && cd btcgo
```

Atualize o arquivo .env com as informações exitentes em .env.example:
```
Número de CPU a ser utilizada
CPU=4

Qual carteira deseja buscar
WALLET=66

Defina o modo de busca. 
  1 - Desde o início
  2 - Sequencial (chave do arquivo)
  3 - Modo aleatório
MODE=3

Deseja utilizar base de dados para controlar as repetições
  1 - Sim
  2 - Não
DB=1

Numero registos por cada busca aleatória.
RANDON_REGISTRY=10000
```

 * Sobre o projeto:
```bash
make up
```

 * Para a execução:
```bash
make down
```

 * Reinicia a aplicação:
```bash
make down
```

 * Atualiza o projeto:
```bash
make update_version
```

 * Acompanhar em tempo real:
```bash
make logs
```

Este container será deletado se a aplicação parar, para executar novamente basta executar o último comando acima.


[install-go]: https://go.dev/doc/install
[install-git]: https://git-scm.com/download/win
[install-docker]: https://www.docker.com/get-started/
[install-docker-compose]: https://docs.docker.com/compose/install/
