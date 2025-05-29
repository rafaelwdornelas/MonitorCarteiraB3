package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	tv "github.com/VictorVictini/tradingview-lib"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
)

// Estruturas para dados da API
type APIResponse struct {
	Data  []Transaction `json:"data"`
	Total int           `json:"total"`
}

type Transaction struct {
	ID             int    `json:"id"`
	TypeInvestment string `json:"type_investment"`
	Ticker         string `json:"ticker"`
	Type           string `json:"type"`
	Date           string `json:"date"`
	Qty            string `json:"qty"`
	Price          string `json:"price_adjusted"`
	Total          string `json:"total"`
	Source         string `json:"source"`
}

// Estrutura para representar um lote de compra
type Lot struct {
	Date     time.Time
	Quantity float64
	Price    float64
}

// Estrutura para armazenar dados do ativo
type Asset struct {
	Ticker         string  `json:"ticker"`
	TypeInvestment string  `json:"typeInvestment"`
	Quantity       float64 `json:"quantity"`
	AveragePrice   float64 `json:"averagePrice"`
	TotalInvested  float64 `json:"totalInvested"`
	CurrentPrice   float64 `json:"currentPrice"`
	CurrentTotal   float64 `json:"currentTotal"`
	ProfitLoss     float64 `json:"profitLoss"`
	ProfitLossPerc float64 `json:"profitLossPerc"`
	mu             sync.RWMutex
}

// Configurações da aplicação
type Config struct {
	Port            string
	Host            string
	Investidor10ID  string
	Investidor10URL string
}

// Gerenciador de portfolio
type PortfolioManager struct {
	assets      map[string]*Asset
	mu          sync.RWMutex
	tradingView *tv.API
	wsClients   map[*websocket.Conn]bool
	wsMu        sync.RWMutex
	config      *Config
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Permite qualquer origem em desenvolvimento
	},
}

func loadConfig() *Config {
	// Carrega variáveis de ambiente do arquivo .env
	err := godotenv.Load()
	if err != nil {
		log.Println("Arquivo .env não encontrado, usando variáveis de ambiente do sistema")
	}

	config := &Config{
		Port:            getEnv("PORT", "4000"),
		Host:            getEnv("HOST", "localhost"),
		Investidor10ID:  getEnv("INVESTIDOR10_ID", "1399345"),
		Investidor10URL: getEnv("INVESTIDOR10_URL", "https://investidor10.com.br/api/carteiras/lancamentos"),
	}

	return config
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func NewPortfolioManager(config *Config) *PortfolioManager {
	return &PortfolioManager{
		assets:    make(map[string]*Asset),
		wsClients: make(map[*websocket.Conn]bool),
		config:    config,
	}
}

// Função para converter string de preço para float64
func parsePrice(price string) float64 {
	// Remove "R$" e espaços
	cleaned := strings.ReplaceAll(price, "R$", "")
	cleaned = strings.ReplaceAll(cleaned, " ", "")
	cleaned = strings.TrimSpace(cleaned)

	// No formato brasileiro: 1.234,56 (ponto para milhar, vírgula para decimal)
	// Remove pontos de milhar
	cleaned = strings.ReplaceAll(cleaned, ".", "")
	// Troca vírgula por ponto para decimal
	cleaned = strings.ReplaceAll(cleaned, ",", ".")

	var value float64
	_, err := fmt.Sscanf(cleaned, "%f", &value)
	if err != nil {
		log.Printf("Erro ao parsear preço '%s': %v", price, err)
	}
	return value
}

// Função para parsear data no formato DD/MM/YYYY
func parseDate(dateStr string) time.Time {
	parts := strings.Split(dateStr, "/")
	if len(parts) != 3 {
		return time.Time{}
	}

	day := 0
	month := 0
	year := 0

	fmt.Sscanf(parts[0], "%d", &day)
	fmt.Sscanf(parts[1], "%d", &month)
	fmt.Sscanf(parts[2], "%d", &year)

	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
}

// Calcula posição usando FIFO
func calculateFIFOPosition(transactions []Transaction) (quantity float64, averagePrice float64, totalInvested float64) {
	// Debug: mostra todas as transações
	log.Printf("\nDEBUG - Processando %d transações", len(transactions))

	// Separa compras e vendas
	var purchases []Lot
	var totalSold float64

	for _, tx := range transactions {
		qty := parsePrice(tx.Qty)
		price := parsePrice(tx.Price)
		total := parsePrice(tx.Total)
		date := parseDate(tx.Date)

		// Debug: mostra valores parseados
		log.Printf("DEBUG - %s %s: Qty=%s (%.2f), Price=%s (%.2f), Total=%s (%.2f)",
			tx.Type, tx.Date, tx.Qty, qty, tx.Price, price, tx.Total, total)

		// Calcula o preço unitário correto (Total / Qty)
		if qty > 0 && total > 0 {
			price = total / qty
			log.Printf("DEBUG - Preço unitário calculado: R$ %.2f", price)
		}

		if tx.Type == "Compra" {
			purchases = append(purchases, Lot{
				Date:     date,
				Quantity: qty,
				Price:    price,
			})
		} else if tx.Type == "Venda" {
			totalSold += qty
		}
	}

	// Ordena compras por data (FIFO - mais antigas primeiro)
	sort.Slice(purchases, func(i, j int) bool {
		return purchases[i].Date.Before(purchases[j].Date)
	})

	log.Printf("DEBUG - Total comprado: %.2f unidades", func() float64 {
		sum := 0.0
		for _, p := range purchases {
			sum += p.Quantity
		}
		return sum
	}())
	log.Printf("DEBUG - Total vendido: %.2f unidades", totalSold)

	// Aplica vendas usando FIFO
	remainingSold := totalSold
	var remainingLots []Lot

	for _, lot := range purchases {
		if remainingSold == 0 {
			remainingLots = append(remainingLots, lot)
		} else if remainingSold >= lot.Quantity {
			// Lote totalmente vendido
			remainingSold -= lot.Quantity
		} else {
			// Lote parcialmente vendido
			remainingLots = append(remainingLots, Lot{
				Date:     lot.Date,
				Quantity: lot.Quantity - remainingSold,
				Price:    lot.Price,
			})
			remainingSold = 0
		}
	}

	// Calcula quantidade, valor investido e preço médio dos lotes restantes
	for _, lot := range remainingLots {
		quantity += lot.Quantity
		totalInvested += lot.Quantity * lot.Price
	}

	if quantity > 0 {
		averagePrice = totalInvested / quantity
	}

	log.Printf("DEBUG - Quantidade restante: %.2f, Preço médio: R$ %.2f, Total investido: R$ %.2f\n",
		quantity, averagePrice, totalInvested)

	return quantity, averagePrice, totalInvested
}

// Busca dados da API
func (pm *PortfolioManager) fetchPortfolioData() error {
	url := fmt.Sprintf("%s/%s/1?draw=1", pm.config.Investidor10URL, pm.config.Investidor10ID)
	log.Printf("Buscando dados de: %s", url)

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var apiResp APIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return err
	}

	// Debug: mostra primeiro registro para ver estrutura
	if len(apiResp.Data) > 0 {
		log.Printf("DEBUG - Primeiro registro: %+v", apiResp.Data[0])
		log.Printf("DEBUG - Price: %s, Total: %s", apiResp.Data[0].Price, apiResp.Data[0].Total)
	}

	// Agrupa transações por ticker
	transactions := make(map[string][]Transaction)
	for _, tx := range apiResp.Data {
		transactions[tx.Ticker] = append(transactions[tx.Ticker], tx)
	}

	// Processa transações agrupadas
	pm.mu.Lock()
	defer pm.mu.Unlock()

	for ticker, txs := range transactions {
		// Calcula posição usando FIFO
		quantity, averagePrice, totalInvested := calculateFIFOPosition(txs)

		// Se vendeu tudo, não adiciona ao portfolio
		if quantity <= 0 {
			log.Printf("%s %s: Posição encerrada", txs[0].TypeInvestment, ticker)
			continue
		}

		pm.assets[ticker] = &Asset{
			Ticker:         ticker,
			TypeInvestment: txs[0].TypeInvestment,
			Quantity:       quantity,
			AveragePrice:   averagePrice,
			TotalInvested:  totalInvested,
		}

		log.Printf("%s %s: %.2f unidades @ R$ %.2f (investido: R$ %.2f)",
			txs[0].TypeInvestment, ticker, quantity, averagePrice, totalInvested)
	}

	return nil
}

// Conecta ao TradingView e monitora preços
func (pm *PortfolioManager) startPriceMonitoring() error {
	pm.tradingView = &tv.API{}
	pm.tradingView.Channels.Read = make(chan map[string]interface{}, 100)
	pm.tradingView.Channels.Error = make(chan error, 10)

	if err := pm.tradingView.OpenConnection(nil); err != nil {
		return err
	}

	// Aguarda conexão
	time.Sleep(2 * time.Second)

	// Adiciona todos os símbolos para monitoramento
	symbols := []string{}
	pm.mu.RLock()
	for ticker, asset := range pm.assets {
		// Para ações, usa BMFBOVESPA:
		// Para ETFs e FIIs, tenta sem prefixo primeiro
		if asset.TypeInvestment == "Ações" {
			symbols = append(symbols, fmt.Sprintf("BMFBOVESPA:%s", ticker))
		} else {
			// Para ETFs e FIIs, adiciona ambas as variações
			symbols = append(symbols, ticker)
			symbols = append(symbols, fmt.Sprintf("BMFBOVESPA:%s", ticker))
		}
	}
	pm.mu.RUnlock()

	if len(symbols) > 0 {
		log.Printf("Monitorando %d símbolos", len(symbols))
		if err := pm.tradingView.AddRealtimeSymbols(symbols); err != nil {
			return err
		}
	}

	// Processa atualizações de preço
	go func() {
		for {
			select {
			case data := <-pm.tradingView.Channels.Read:
				pm.processPriceUpdate(data)
			case err := <-pm.tradingView.Channels.Error:
				log.Printf("Erro TradingView: %v", err)
			}
		}
	}()

	return nil
}

// Processa atualização de preço
func (pm *PortfolioManager) processPriceUpdate(data map[string]interface{}) {
	symbolField, ok := data["symbol"].(string)
	if !ok {
		return
	}

	// Remove prefixo BMFBOVESPA: se existir
	ticker := symbolField
	if len(ticker) > 11 && ticker[:11] == "BMFBOVESPA:" {
		ticker = ticker[11:]
	}

	price, ok := data["current_price"].(float64)
	if !ok || price == 0 {
		return
	}

	pm.mu.RLock()
	asset, exists := pm.assets[ticker]
	pm.mu.RUnlock()

	if !exists {
		return
	}

	// Atualiza preço e calcula ganhos
	asset.mu.Lock()
	oldPrice := asset.CurrentPrice
	asset.CurrentPrice = price
	asset.CurrentTotal = asset.Quantity * price
	asset.ProfitLoss = asset.CurrentTotal - asset.TotalInvested
	if asset.TotalInvested > 0 {
		asset.ProfitLossPerc = (asset.ProfitLoss / asset.TotalInvested) * 100
	}
	asset.mu.Unlock()

	// Log apenas se o preço mudou
	if oldPrice != price {
		log.Printf("%s atualizado: R$ %.2f -> R$ %.2f", ticker, oldPrice, price)
	}

	// Envia atualização para clientes WebSocket
	pm.broadcastUpdate(asset)
}

// Envia atualização para todos os clientes WebSocket
func (pm *PortfolioManager) broadcastUpdate(asset *Asset) {
	asset.mu.RLock()
	data, _ := json.Marshal(asset)
	asset.mu.RUnlock()

	pm.wsMu.RLock()
	defer pm.wsMu.RUnlock()

	for client := range pm.wsClients {
		if err := client.WriteMessage(websocket.TextMessage, data); err != nil {
			client.Close()
			delete(pm.wsClients, client)
		}
	}
}

// Handler WebSocket
func (pm *PortfolioManager) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Erro ao fazer upgrade WebSocket: %v", err)
		return
	}
	defer conn.Close()

	pm.wsMu.Lock()
	pm.wsClients[conn] = true
	pm.wsMu.Unlock()

	log.Println("Novo cliente WebSocket conectado")

	// Envia dados iniciais
	pm.mu.RLock()
	for _, asset := range pm.assets {
		asset.mu.RLock()
		data, _ := json.Marshal(asset)
		asset.mu.RUnlock()
		conn.WriteMessage(websocket.TextMessage, data)
	}
	pm.mu.RUnlock()

	// Mantém conexão aberta
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			pm.wsMu.Lock()
			delete(pm.wsClients, conn)
			pm.wsMu.Unlock()
			break
		}
	}
}

// Handler para página principal
func (pm *PortfolioManager) handleIndex(w http.ResponseWriter, r *http.Request) {
	// Serve o arquivo HTML
	http.ServeFile(w, r, "index.html")
}

// Handler para API de resumo da carteira
func (pm *PortfolioManager) handlePortfolioSummary(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	pm.mu.RLock()
	defer pm.mu.RUnlock()

	summary := struct {
		TotalAssets   int      `json:"totalAssets"`
		TotalInvested float64  `json:"totalInvested"`
		CurrentTotal  float64  `json:"currentTotal"`
		TotalProfit   float64  `json:"totalProfit"`
		ProfitPercent float64  `json:"profitPercent"`
		Assets        []*Asset `json:"assets"`
	}{}

	for _, asset := range pm.assets {
		asset.mu.RLock()
		summary.TotalAssets++
		summary.TotalInvested += asset.TotalInvested
		summary.CurrentTotal += asset.CurrentTotal
		summary.TotalProfit += asset.ProfitLoss
		summary.Assets = append(summary.Assets, asset)
		asset.mu.RUnlock()
	}

	if summary.TotalInvested > 0 {
		summary.ProfitPercent = (summary.TotalProfit / summary.TotalInvested) * 100
	}

	json.NewEncoder(w).Encode(summary)
}

func main() {
	// Carrega configurações
	config := loadConfig()

	log.Printf("=== Configurações ===")
	log.Printf("Host: %s", config.Host)
	log.Printf("Porta: %s", config.Port)
	log.Printf("ID Investidor10: %s", config.Investidor10ID)
	log.Printf("====================\n")

	pm := NewPortfolioManager(config)

	// Busca dados da carteira
	log.Println("Buscando dados da carteira...")
	if err := pm.fetchPortfolioData(); err != nil {
		log.Fatalf("Erro ao buscar dados: %v", err)
	}

	// Exibe resumo da carteira
	pm.mu.RLock()
	totalInvested := 0.0
	totalAssets := len(pm.assets)
	for _, asset := range pm.assets {
		totalInvested += asset.TotalInvested
	}
	pm.mu.RUnlock()

	log.Printf("\n=== RESUMO DA CARTEIRA ===")
	log.Printf("Total de ativos: %d", totalAssets)
	log.Printf("Total investido: R$ %.2f", totalInvested)
	log.Printf("========================\n")

	// Inicia monitoramento de preços
	log.Println("Iniciando monitoramento de preços...")
	if err := pm.startPriceMonitoring(); err != nil {
		log.Fatalf("Erro ao iniciar monitoramento: %v", err)
	}

	// Configura rotas
	r := mux.NewRouter()
	r.HandleFunc("/", pm.handleIndex)
	r.HandleFunc("/ws", pm.handleWebSocket)
	r.HandleFunc("/api/portfolio", pm.handlePortfolioSummary).Methods("GET")

	// Serve arquivos estáticos (caso precise de CSS/JS separados no futuro)
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))

	// Inicia servidor
	addr := fmt.Sprintf(":%s", config.Port)
	log.Printf("Servidor rodando em http://%s%s", config.Host, addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("Erro ao iniciar servidor: %v", err)
	}
}
