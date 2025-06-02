# Monitor de Carteira B3

[![Go Version](https://img.shields.io/badge/Go-1.23+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

Monitor em tempo real para carteiras de investimentos da B3, integrando dados do Investidor10 com cotações ao vivo do TradingView e notícias relevantes dos seus ativos.

## 📊 Características

### Monitoramento em Tempo Real
- **Cotações ao Vivo**: Acompanhe os preços dos seus ativos em tempo real via TradingView
- **Indicadores Visuais**: Animações em verde (alta) e vermelho (baixa) nas mudanças de preço
- **WebSocket**: Atualizações instantâneas sem necessidade de recarregar a página
- **Status de Conexão**: Indicador visual do estado da conexão em tempo real

### Gestão de Portfolio
- **Integração Investidor10**: Importa automaticamente suas transações e posições
- **Cálculo FIFO**: Utiliza o método FIFO (First In, First Out) para calcular preço médio
- **Organização por Tipo**: Separa automaticamente Ações, ETFs, FIIs e outros investimentos
- **Seções Recolhíveis**: Interface organizada com seções que podem ser expandidas/recolhidas
- **Ordenação Dinâmica**: Clique nos cabeçalhos das colunas para ordenar os dados

### Sistema de Notícias 🆕
- **Monitoramento de Notícias**: Verifica notícias relevantes dos seus ativos a cada minuto
- **Alertas em Tempo Real**: Notificações visuais e sonoras para novas notícias
- **Filtro Inteligente**: Mostra apenas notícias do dia atual em português
- **Urgência Destacada**: Notícias urgentes são destacadas visualmente
- **Link Direto**: Acesso rápido às notícias no TradingView

### Cálculos Automáticos
- **Lucro/Prejuízo**: Em reais e percentual por ativo e total
- **Preço Médio**: Calculado considerando todas as transações
- **Performance por Setor**: Totais separados por tipo de investimento
- **Rentabilidade Global**: Visão geral da performance da carteira

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

### Elementos da Interface

#### Dashboard Principal
- **Cards de Resumo**: Total investido, valor atual, lucro/prejuízo e rentabilidade
- **Status de Conexão**: Indicador verde (conectado) ou vermelho (desconectado)

#### Seções de Investimentos
- **Cabeçalho Clicável**: Clique para expandir/recolher cada seção
- **Estatísticas da Seção**: Total investido, atual e lucro por tipo de ativo
- **Tabela de Ativos**: 
  - Link direto para o gráfico no TradingView
  - Quantidade e preço médio de compra
  - Valores atuais em tempo real
  - Indicadores coloridos de performance

#### Sistema de Notificações
- **Pop-ups de Notícias**: Aparecem no canto superior direito
- **Indicador de Urgência**: Destaque para notícias importantes
- **Auto-dismiss**: Notificações somem após 30 segundos
- **Botão de Fechar**: Opção para fechar manualmente

## 🛠️ Tecnologias Utilizadas

### Backend
- **Go 1.23**: Linguagem principal com suporte a goroutines
- **Gorilla Mux**: Roteamento HTTP robusto
- **Gorilla WebSocket**: Comunicação bidirecional em tempo real
- **TradingView Lib**: Integração com cotações ao vivo
- **Sync.Mutex**: Controle de concorrência thread-safe

### Frontend
- **HTML5/CSS3**: Interface moderna e responsiva
- **JavaScript Vanilla**: Sem dependências externas, código otimizado
- **WebSocket API**: Comunicação em tempo real
- **Web Audio API**: Sons de notificação nativos
- **CSS Animations**: Feedback visual suave

## 📦 Estrutura do Projeto

```
MonitorCarteiraB3/
├── main.go           # Aplicação principal com toda lógica
├── index.html        # Interface web completa
├── go.mod           # Dependências Go
├── go.sum           # Checksums das dependências
├── .env             # Configurações (não versionado)
├── .gitignore       # Arquivos ignorados pelo Git
└── README.md        # Este arquivo
```

## 📈 Como Funciona

### 1. Inicialização
- Carrega configurações do arquivo `.env`
- Busca transações do Investidor10
- Calcula posições usando método FIFO

### 2. Monitoramento de Preços
- Conecta ao TradingView via WebSocket
- Monitora todos os símbolos da carteira
- Envia atualizações para clientes conectados

### 3. Sistema de Notícias
- Verifica notícias a cada minuto para cada ativo
- Filtra notícias do dia em português
- Envia alertas apenas para notícias novas
- Mantém histórico para evitar duplicatas

### 4. Interface Web
- Conecta via WebSocket ao servidor
- Recebe atualizações de preços e notícias
- Atualiza DOM dinamicamente
- Mantém estado de ordenação e seções

## 🔧 API Endpoints

### WebSocket
- `ws://localhost:4000/ws` - Conexão principal para atualizações

### HTTP
- `GET /` - Interface web principal
- `GET /api/portfolio` - Resumo da carteira em JSON
- `GET /api/news/recent` - Informações sobre notícias recentes

## 🎯 Recursos Avançados

### Cálculo FIFO
O sistema implementa o método FIFO para calcular o preço médio correto considerando vendas parciais:
- Ordena compras por data
- Aplica vendas às compras mais antigas primeiro
- Recalcula preço médio apenas dos lotes restantes

### Detecção de Mudanças
- Compara preços anteriores com atuais
- Aplica animação verde ou vermelha conforme direção
- Atualiza apenas elementos modificados (otimização)

### Gestão de Conexões
- Reconexão automática em caso de queda
- Múltiplos clientes simultâneos suportados
- Limpeza automática de conexões inativas

## 🤝 Contribuindo

Contribuições são bem-vindas! Sinta-se à vontade para:

1. Fazer um Fork do projeto
2. Criar uma branch para sua feature (`git checkout -b feature/AmazingFeature`)
3. Commit suas mudanças (`git commit -m 'Add some AmazingFeature'`)
4. Push para a branch (`git push origin feature/AmazingFeature`)
5. Abrir um Pull Request

### Sugestões de Melhorias
- [ ] Suporte a múltiplas carteiras
- [ ] Gráficos de evolução histórica
- [ ] Exportação de relatórios
- [ ] Modo escuro/claro
- [ ] App mobile

## 📝 Licença

Este projeto está sob a licença MIT. Veja o arquivo [LICENSE](LICENSE) para mais detalhes.

## 🐛 Problemas Conhecidos

- Alguns ETFs e FIIs podem não ter cotações disponíveis no TradingView
- A API do Investidor10 pode ter limitações de taxa de requisição
- Notificações sonoras requerem interação inicial do usuário (limitação do navegador)

## 🔒 Segurança

- Todas as comunicações são locais por padrão
- Não armazena senhas ou dados sensíveis
- ID do Investidor10 é usado apenas para leitura

## 📞 Suporte

Para reportar bugs ou sugerir melhorias, abra uma [issue](https://github.com/rafaelwdornelas/MonitorCarteiraB3/issues).

## 🙏 Agradecimentos

- [Investidor10](https://investidor10.com.br) pela API de dados
- [TradingView](https://tradingview.com) pelas cotações e notícias em tempo real
- [VictorVictini](https://github.com/VictorVictini) pela biblioteca TradingView
- Comunidade Go pela excelente documentação

---

Desenvolvido com ❤️ por [Rafael W. Dornelas](https://github.com/rafaelwdornelas)
