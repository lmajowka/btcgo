# BTC GO v0.55 - Fork by MartonLyra

# Instruções para rodar o projeto (na sua máquina)

[![instalação do Go no Windows](https://img.youtube.com/vi/679Zc7ZQLtI/0.jpg)](https://www.youtube.com/watch?v=679Zc7ZQLtI)

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
   ``` docker buildx build --no-cache -t btcgo .```
 * Executa a imagem contruída no passo anterior:
   ``` docker run -it --name btcgo btcgo```



[install-go]: https://go.dev/doc/install
[install-docker]: https://www.docker.com/get-started/
[install-docker-compose]: https://docs.docker.com/compose/install/

# Diferenciais desse Fork

### Modo 3 - Aleatório
* No Modo 3 - Aleatório, de 2 em 2 horas (tempo definido em tickerTime2randomAddress), o sistema sorteira uma nova posição dentro do range da carteira escolhida;
* Caso a nova posição esteja no início da carteira (primeiros 4%) ou ao final da carteira (últimos 1%), uma nova posição é sorteada aleatoriamente
* Quando uma nova posição é escolhida, sua posição em forma gráfica é exibida;
* Sempre checamos se a chave sendo buscada está dentro do range da carteira. Não queremos perder tempo pesquisando fora do range.

### Exibição com mais detalhes
* A cada 5 segundos, as informações da busca são melhores detalhados, conforme exemplo abaixo:
``` 2024-07-01 12:33:21 - Posição: 0x3e129edc38f6c1781 (92.977301609240%) ; Chaves checadas: 235,600,258 ; Chaves por segundo: 795,614 ; Tempo restante: 1012429 anos```


*Caso o autor conceda permissão, posso implementar essas e outras alterações no repositório original*