# Monitor de Carteira B3

[![Go Version](https://img.shields.io/badge/Go-1.23+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

Monitor em tempo real para carteiras de investimentos da B3, integrando dados do Investidor10 com cotações ao vivo do TradingView.

## 📊 Características

- **Monitoramento em Tempo Real**: Acompanhe os preços dos seus ativos ao vivo
- **Integração com Investidor10**: Importa automaticamente suas transações e posições
- **Interface Web Responsiva**: Dashboard moderno e intuitivo
- **WebSocket**: Atualizações instantâneas sem necessidade de recarregar a página
- **Cálculo Automático**: Lucro/prejuízo, percentuais e totais calculados automaticamente
- **Organização por Tipo**: Separa automaticamente Ações, ETFs, FIIs e outros investimentos

## 🚀 Instalação

### Pré-requisitos

- Go 1.23 ou superior
- Conta no [Investidor10](https://investidor10.com.br)

### Clone o repositório

```bash
git clone https://github.com/rafaelwdornelas/MonitorCarteiraB3.git
cd MonitorCarteiraB3
```

### Configure as variáveis de ambiente

Crie um arquivo `.env` na raiz do projeto:

```env
# Configurações do servidor
PORT=4000
HOST=127.0.0.1

# Configurações do Investidor10
INVESTIDOR10_ID=SEU_ID_AQUI

# Configurações do TradingView (opcional)
TRADINGVIEW_RECONNECT_INTERVAL=5
TRADINGVIEW_MAX_RETRIES=10
```

### Como obter seu ID do Investidor10

1. Acesse sua carteira no [Investidor10](https://investidor10.com.br)
2. Na URL, você verá algo como: `https://investidor10.com.br/carteiras/1399345`
3. O número após `/carteiras/` é seu ID (no exemplo: 1399345)

### Instale as dependências

```bash
go mod download
```

### Execute o projeto

```bash
go run main.go
```

Acesse http://localhost:4000 no seu navegador.

## 📱 Interface

![Monitor de Carteira B3](screenshot.png)

A interface mostra:

- **Resumo Geral**: Total investido, valor atual, lucro/prejuízo e rentabilidade
- **Seções por Tipo**: Ações, ETFs, FIIs organizados separadamente
- **Detalhes por Ativo**: 
  - Quantidade e preço médio
  - Total investido e valor atual
  - Lucro/prejuízo em R$ e %
- **Status de Conexão**: Indicador visual do estado da conexão WebSocket

## 🛠️ Tecnologias Utilizadas

### Backend
- **Go**: Linguagem principal
- **Gorilla Mux**: Roteamento HTTP
- **Gorilla WebSocket**: Comunicação em tempo real
- **TradingView Lib**: Integração com cotações ao vivo

### Frontend
- **HTML5/CSS3**: Interface responsiva
- **JavaScript Vanilla**: Sem dependências externas
- **WebSocket**: Atualizações em tempo real

## 📦 Estrutura do Projeto

```
MonitorCarteiraB3/
├── main.go           # Aplicação principal
├── index.html        # Interface web
├── go.mod           # Dependências Go
├── go.sum           # Checksums das dependências
├── .env             # Configurações (não versionado)
└── README.md        # Este arquivo
```

## 📈 Como Funciona

1. **Importação de Dados**: O sistema busca suas transações do Investidor10
2. **Cálculo de Posições**: Agrupa transações por ativo e calcula médias
3. **Monitoramento de Preços**: Conecta ao TradingView para cotações em tempo real
4. **Atualização da Interface**: Envia atualizações via WebSocket para o navegador

## 🤝 Contribuindo

Contribuições são bem-vindas! Sinta-se à vontade para:

1. Fazer um Fork do projeto
2. Criar uma branch para sua feature (`git checkout -b feature/AmazingFeature`)
3. Commit suas mudanças (`git commit -m 'Add some AmazingFeature'`)
4. Push para a branch (`git push origin feature/AmazingFeature`)
5. Abrir um Pull Request

## 📝 Licença

Este projeto está sob a licença MIT. Veja o arquivo [LICENSE](LICENSE) para mais detalhes.

## 🐛 Problemas Conhecidos

- Alguns ETFs e FIIs podem não ter cotações disponíveis no TradingView
- A API do Investidor10 pode ter limitações de taxa de requisição

## 📞 Suporte

Para reportar bugs ou sugerir melhorias, abra uma [issue](https://github.com/rafaelwdornelas/MonitorCarteiraB3/issues).

## 🙏 Agradecimentos

- [Investidor10](https://investidor10.com.br) pela API de dados
- [TradingView](https://tradingview.com) pelas cotações em tempo real
- [VictorVictini](https://github.com/VictorVictini) pela biblioteca TradingView

---

Desenvolvido com ❤️ por [Rafael W. Dornelas](https://github.com/rafaelwdornelas)
