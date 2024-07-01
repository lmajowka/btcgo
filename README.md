# BTC GO v0.5

# Instruções para rodar o projeto (na sua máquina)

[![instalação do Go no Windows](https://img.youtube.com/vi/679Zc7ZQLtI/0.jpg)](https://www.youtube.com/watch?v=679Zc7ZQLtI)

## Requisitos
  -  [Go][install-go]
  -  Terminal

## Execução do corre
Se liga no esquema pra rodar o bagulho:

  Copiar o arquivo ``` dev.env ``` para ``` .env ``` e alterar os valores para o seu ambiente.

 * Para ambientes windows:

  Subir ambiente: 
  ``` make windows-up ```

  Derrubar ambiente: 
  ``` make windows-down ```

 * Para ambientes linux:

  Subir ambiente: 
  ``` make linux-up ```

  Derrubar ambiente: 
  ``` make linux-down ```

 
# Instruções para rodar o projeto (em container)

## Requisitos
  -  [Docker][install-docker]
  -  [Docker-compose][install-docker-compose]

## Execução da parada
É tão fácil como voar:

``` make docker-up```

Para parar a aplicacao, basta executar:

``` make docker-down ```


[install-go]: https://go.dev/doc/install
[install-docker]: https://www.docker.com/get-started/
[install-docker-compose]: https://docs.docker.com/compose/install/