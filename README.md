# BTC GO v0.6.0
[![instalação do Go no Windows](https://img.youtube.com/vi/679Zc7ZQLtI/0.jpg)](https://www.youtube.com/watch?v=679Zc7ZQLtI)

## Requisitos
  -  [Go][install-go]
  -  [Git][install-git]
  -  Terminal

# Instruções para rodar o projeto (na Windows)

 * Clona o repo e entra na pasta:
```bash
  git clone git@github.com:lmajowka/btcgo.git && cd btcgo
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


# Instruções para rodar o projeto (no Linux ou MacOS)

 * Clona o repo e entra na pasta:
```bash
  git clone git@github.com:lmajowka/btcgo.git && cd btcgo
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

# Instruções para rodar o projeto (em container)

## Requisitos
  -  [Docker][install-docker]
  -  [Docker-compose][install-docker-compose]
  -  [Git][install-git]

## Execução da parada
É tão fácil como voar:

 * Clona o repo:
  ``` git clone git@github.com:lmajowka/btcgo.git && cd btcgo```
 * Build do Dockerfile:
   ``` docker buildx build --no-cache -t btcgo .```
 * Executa a imagem contruída no passo anterior:
   ``` docker run --rm -it --name btcgo btcgo```

Este container será deletado se a aplicação for parada, para  executar novamente posteriormente, basta executar o último comando acima.


[install-go]: https://go.dev/doc/install
[install-git]: https://git-scm.com/download/win
[install-docker]: https://www.docker.com/get-started/
[install-docker-compose]: https://docs.docker.com/compose/install/