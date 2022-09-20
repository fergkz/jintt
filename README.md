# jintt

![GitHub repo size](https://img.shields.io/github/repo-size/fergkz/jintt?style=for-the-badge)
![GitHub language count](https://img.shields.io/github/languages/count/fergkz/jintt?style=for-the-badge)
![GitHub forks](https://img.shields.io/github/forks/fergkz/jintt?style=for-the-badge)
![Bitbucket open issues](https://img.shields.io/bitbucket/issues/fergkz/jintt?style=for-the-badge)
![Bitbucket open pull requests](https://img.shields.io/bitbucket/pr-raw/fergkz/jintt?style=for-the-badge)


> Extraia um relatório das atividades e alocações da sua sprint no jira

## 🏠 Ajustes e melhorias

O projeto ainda está em desenvolvimento

## 💻 Pré-requisitos

Antes de começar, verifique se você atendeu aos seguintes requisitos:

* Você instalou no mínimo a versão do `go 1.16`
* Você tem uma máquina `Windows / Linux / Mac`

## 🚀 Instalando o `jintt`

Para instalar o `jintt`, siga estas etapas:

* Instale o `go 1.16`
* Clone o repositório na sua máquina
* Instale os pacotes adicionais (ex: `go mod tidy`)
* Copie o arquivo `config-example.yml` e renomeie para `config.yml`
* Preencha as informações do arquivo `config.yml` com as suas configurações

## ☕ Compilando o `jintt`

Você pode compilar este serviço para rodar diretamente na sua máquina, basta executar os comandos:

### Linux
```
> $Env:GOOS = "linux"; $Env:GOARCH = "amd64"
> go build -o launcher .
```

### Windows
```
> $Env:GOOS = "windows"; $Env:GOARCH = "amd64"
> go build -o launcher.exe .
```


## 🏁 Usando o `jintt`

### Rodando o `jintt` compilado:

* Basta executar o arquivo `laucher` no linux ou `launcher.exe` no windows
* Acesse do seu navegador a url `http://localhost/sprint/{NÚMERO_DA_SPRINT}`

### Rodando o `jintt` em desenvolvimento:

* Execute o comando `go run .`
* Acesse do seu navegador a url `http://localhost/sprint/{NÚMERO_DA_SPRINT}`


## 😄 Seja uma das pessoas contribuidoras

Para contribuir com `jintt`, siga estas etapas:

1. Bifurque este repositório.
2. Crie um branch: `git checkout -b main`.
3. Faça suas alterações e confirme-as: `git commit -m '<mensagem_commit>'`
4. Envie para o branch original: `git push origin jintt/main`
5. Crie a solicitação de pull.

Como alternativa, consulte a documentação do GitHub em [como criar uma solicitação pull](https://help.github.com/en/github/collaborating-with-issues-and-pull-requests/creating-a-pull-request).


## 📝 Licença

Esse projeto está sob licença. Veja o arquivo [LICENÇA](LICENSE.md) para mais detalhes.

[⬆ Voltar ao topo](#jintt)<br>