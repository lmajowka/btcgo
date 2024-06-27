# BTC GO v0.3

# Instruções para rodar o projeto (na sua máquina)

## Requisitos
  -  [Go][install-go]
  -  Terminal

## Execução do corre
Se liga no esquema pra rodar o bagulho:

 * Clona o repo:
  ``` git clone git@github.com:lmajowka/btcgo.git ```
 * Brota na pasta do projeto:
  ``` cd btcgo ```
 * Instala as parada:
 ``` go mod tidy ```
 * Faz o build do projeto no LINUX:
 ``` go build -o btcgo ./src ``` 

  * Faz o build do projeto no WINDOWS:
 ``` go build -o btcgo.exe ./src ``` 
 * Executa o que foi compilado:
 ``` ./btcgo ```

Aí é só seguir o baile, parceiro.

# Instruções para rodar o projeto (em container)

## Requisitos
  -  [Docker][install-docker]
  -  [Docker-compose][install-docker-compose]

## Execução da parada
É tão fácil como voar:

 * Clona o repo:
  ``` git clone git@github.com:lmajowka/btcgo.git && cd btcgo```
 * Build do Dockerfile:
   ``` docker buildx build --no-cache -t btcgo```
 * Executa a imagem contruída no passo anterior:
   ``` docker run -it --name btcgo btcgo```



[install-go]: https://www.docker.com/get-started/
[install-docker]: https://www.docker.com/get-started/
[install-docker-compose]: https://docs.docker.com/compose/install/