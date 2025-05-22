package main

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Define types for our genesis file.
type GenesisAccount struct {
	Address       string  `json:"address"`
	AccountNumber string  `json:"account_number"`
	Sequence      string  `json:"sequence"`
	PubKey        *PubKey `json:"pub_key,omitempty"`
}

type PubKey struct {
	Type  string `json:"@type"`
	Value string `json:"key"`
}

type Coin struct {
	Denom  string `json:"denom"`
	Amount string `json:"amount"`
}

type Balance struct {
	Address string `json:"address"`
	Coins   []Coin `json:"coins"`
}

type AuthGenesis struct {
	Accounts []GenesisAccount `json:"accounts"`
	Params   AuthParams       `json:"params"`
}

type AuthParams struct {
	MaxMemoCharacters      string `json:"max_memo_characters"`
	TxSigLimit             string `json:"tx_sig_limit"`
	TxSizeCostPerByte      string `json:"tx_size_cost_per_byte"`
	SigVerifyCostEd25519   string `json:"sig_verify_cost_ed25519"`
	SigVerifyCostSecp256k1 string `json:"sig_verify_cost_secp256k1"`
}

type BankGenesis struct {
	Params        BankParams `json:"params"`
	Balances      []Balance  `json:"balances"`
	Supply        []Coin     `json:"supply"`
	DenomMetadata []Metadata `json:"denom_metadata"`
}

type BankParams struct {
	SendEnabled        []SendEnabled `json:"send_enabled"`
	DefaultSendEnabled bool          `json:"default_send_enabled"`
}

type SendEnabled struct {
	Denom   string `json:"denom"`
	Enabled bool   `json:"enabled"`
}

type Metadata struct {
	Description string      `json:"description"`
	DenomUnits  []DenomUnit `json:"denom_units"`
	Base        string      `json:"base"`
	Display     string      `json:"display"`
	Name        string      `json:"name"`
	Symbol      string      `json:"symbol"`
}

type DenomUnit struct {
	Denom    string   `json:"denom"`
	Exponent uint32   `json:"exponent"`
	Aliases  []string `json:"aliases,omitempty"`
}

type GenesisDoc struct {
	GenesisTime     time.Time      `json:"genesis_time"`
	ChainID         string         `json:"chain_id"`
	InitialHeight   string         `json:"initial_height"`
	ConsensusParams map[string]any `json:"consensus_params"`
	AppState        AppState       `json:"app_state"`
}

type AppState struct {
	Auth    AuthGenesis    `json:"auth"`
	Bank    BankGenesis    `json:"bank"`
	Staking map[string]any `json:"staking"`
	Genutil map[string]any `json:"genutil"`
	// Other modules can be added as needed
}

// Holds our internal processing data.
type GenesisData struct {
	Accounts       map[string]*GenesisAccount
	Balances       map[string][]Coin
	Supply         []Coin
	AccountCounter int // Counter for assigning sequential account numbers
}

// Constant for address prefix conversion.
const (
	OldPrefix = "unicorn"
	NewPrefix = "gadikian"
)

func main() {
	err := run()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

// run performs the main logic of the program.
func run() error {
	// Check if IPFS directory is provided
	if len(os.Args) < 2 {
		return errors.New("usage: genesis-tool <ipfs-dir> [chain-id]")
	}

	ipfsDir := os.Args[1]
	chainID := "gadikian-1" // Default chain-id changed to gadikian
	if len(os.Args) > 2 {
		chainID = os.Args[2]
	} else if strings.HasPrefix(chainID, "unicorn") {
		// Make sure we're using the gadikian chain ID if not specified
		chainID = strings.Replace(chainID, "unicorn", "gadikian", 1)
	}

	// Initialize genesis data
	genesisData := &GenesisData{
		Accounts:       make(map[string]*GenesisAccount),
		Balances:       make(map[string][]Coin),
		Supply:         make([]Coin, 0),
		AccountCounter: 0, // Initialize the account counter
	}

	// Process files
	if err := processFiles(ipfsDir, genesisData); err != nil {
		return err
	}

	// Convert prefixes
	fmt.Println("Converting bech32 prefixes from 'unicorn' to 'gadikian'...")
	convertPrefixes(genesisData)

	// Create final genesis.json
	fmt.Println("Generating genesis.json...")
	if err := generateGenesisJSON(genesisData, chainID); err != nil {
		return fmt.Errorf("error generating genesis JSON: %w", err)
	}

	fmt.Println("Genesis file created successfully: genesis.json")
	fmt.Printf("Chain ID: %s\n", chainID)
	fmt.Println("Addresses converted from unicorn prefix to gadikian prefix")
	fmt.Println("Token denoms converted from uwunicorn to ugadikian")
	fmt.Printf("Total accounts created: %d\n", genesisData.AccountCounter)

	return nil
}

// processFiles processes all CSV files and populates the genesis data.
func processFiles(ipfsDir string, data *GenesisData) error {
	// Process balances.csv if it exists
	fmt.Println("Processing balances.csv...")
	balancesPath := filepath.Join(ipfsDir, "balances.csv")
	if _, err := os.Stat(balancesPath); err == nil {
		if err := processBalances(balancesPath, data); err != nil {
			return fmt.Errorf("error processing balances: %w", err)
		}
	} else {
		fmt.Println("balances.csv not found, skipping...")
	}

	// Process supply.csv
	fmt.Println("Processing supply.csv...")
	supplyPath := filepath.Join(ipfsDir, "supply.csv")
	if err := processSupply(supplyPath, data); err != nil {
		return fmt.Errorf("error processing supply: %w", err)
	}

	// Process kaway_bond.csv
	fmt.Println("Processing kaway_bond.csv...")
	kawayPath := filepath.Join(ipfsDir, "kaway_bond.csv")
	if err := processBonds(kawayPath, "uwunicorn", data); err != nil {
		return fmt.Errorf("error processing kaway_bond: %w", err)
	}

	// Process uwuval_bond.csv if it exists
	fmt.Println("Processing uwuval_bond.csv...")
	uwuvalPath := filepath.Join(ipfsDir, "uwuval_bond.csv")
	if _, err := os.Stat(uwuvalPath); err == nil {
		if err := processBonds(uwuvalPath, "valuwunicorn", data); err != nil {
			return fmt.Errorf("error processing uwuval_bond: %w", err)
		}
	} else {
		fmt.Println("uwuval_bond.csv not found, skipping...")
	}

	// Process LP files
	fmt.Println("Processing liquidity pool data...")
	processLPs(ipfsDir)

	return nil
}

// Convert unicorn prefixes to gadikian.
func convertPrefixes(data *GenesisData) {
	// Convert account addresses
	convertedAccounts := make(map[string]*GenesisAccount)
	for oldAddress, account := range data.Accounts {
		newAddress := convertAddress(oldAddress)
		// Also update the address in the account itself
		account.Address = newAddress
		convertedAccounts[newAddress] = account
	}
	data.Accounts = convertedAccounts

	// Convert balances addresses and denoms
	convertedBalances := make(map[string][]Coin)
	for oldAddress, coins := range data.Balances {
		newAddress := convertAddress(oldAddress)

		// Convert denoms in coins
		convertedCoins := make([]Coin, len(coins))
		for i, coin := range coins {
			convertedCoins[i] = Coin{
				Denom:  convertDenom(coin.Denom),
				Amount: coin.Amount,
			}
		}

		convertedBalances[newAddress] = convertedCoins
	}
	data.Balances = convertedBalances

	// Convert supply denoms
	for i, coin := range data.Supply {
		data.Supply[i] = Coin{
			Denom:  convertDenom(coin.Denom),
			Amount: coin.Amount,
		}
	}

	// Debug output
	fmt.Printf("Converted %d account addresses\n", len(data.Accounts))
	fmt.Printf("Converted %d balance entries\n", len(data.Balances))
	fmt.Printf("Converted %d supply entries\n", len(data.Supply))
}

// Convert a unicorn address to a gadikian address.
func convertAddress(address string) string {
	if strings.HasPrefix(address, OldPrefix) {
		return NewPrefix + address[len(OldPrefix):]
	}
	return address
}

// Convert a denom from unicorn-prefixed to gadikian-prefixed.
func convertDenom(denom string) string {
	// Convert uwunicorn -> ugadikian
	if denom == "uwunicorn" {
		return "ugadikian"
	}

	// Convert valuwunicorn -> valgadikian
	if denom == "valuwunicorn" {
		return "valgadikian"
	}

	// Convert factory/unicorn.../uXXX -> factory/gadikian.../uXXX
	if strings.HasPrefix(denom, "factory/"+OldPrefix) {
		return "factory/" + NewPrefix + denom[len("factory/"+OldPrefix):]
	}

	return denom
}

// Process bonds from a CSV file (kaway_bond.csv or uwuval_bond.csv).
func processBonds(filePath, denom string, data *GenesisData) error {
	// Check if file exists
	if _, err := os.Stat(filePath); err != nil {
		fmt.Printf("%s not found, skipping...\n", filepath.Base(filePath))
		return nil
	}

	// Open the CSV file
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open %s: %w", filepath.Base(filePath), err)
	}
	defer file.Close()

	// Create a new CSV reader
	reader := csv.NewReader(file)

	// Read the header
	header, err := reader.Read()
	if err != nil {
		return fmt.Errorf("failed to read header: %w", err)
	}

	// Expected header: address,uwu or similar
	if len(header) < 2 || header[0] != "address" {
		return fmt.Errorf("unexpected header format in %s, expected: address,amount", filepath.Base(filePath))
	}

	return processCsvRows(reader, data, denom)
}

// Process CSV rows to extract bond data.
func processCsvRows(reader *csv.Reader, data *GenesisData, denom string) error {
	// Process rows
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("error reading row: %w", err)
		}

		// Parse data
		address := row[0]
		amount := row[1]

		// Skip empty amounts
		if amount == "" || amount == "0" {
			continue
		}

		// Create coin
		coin := Coin{
			Denom:  denom,
			Amount: amount,
		}

		// Ensure account exists
		if _, exists := data.Accounts[address]; !exists {
			accountNumber := data.AccountCounter
			data.AccountCounter++

			data.Accounts[address] = &GenesisAccount{
				Address:       address,
				AccountNumber: fmt.Sprintf("%d", accountNumber),
				Sequence:      "0",
			}
		}

		// Add coin to balances
		data.Balances[address] = append(data.Balances[address], coin)
	}

	return nil
}

// Process balances.csv - expected to be complex with many columns.
func processBalances(filePath string, data *GenesisData) error {
	// Open the CSV file
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open balances.csv: %w", err)
	}
	defer file.Close()

	// Create a new CSV reader
	reader := csv.NewReader(file)

	// Read the header to determine columns
	header, err := reader.Read()
	if err != nil {
		return fmt.Errorf("failed to read header: %w", err)
	}

	// First column should be 'address'
	if header[0] != "address" {
		return errors.New("unexpected header format in balances.csv, first column should be 'address'")
	}

	return processBalanceRows(reader, header, data)
}

// Process balance rows from CSV.
func processBalanceRows(reader *csv.Reader, header []string, data *GenesisData) error {
	// Process rows
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("error reading row: %w", err)
		}

		if err := processBalanceRow(row, header, data); err != nil {
			return err
		}
	}

	return nil
}

// Process a single balance row.
func processBalanceRow(row []string, header []string, data *GenesisData) error {
	// Parse address
	address := row[0]

	// Ensure account exists
	if _, exists := data.Accounts[address]; !exists {
		accountNumber := data.AccountCounter
		data.AccountCounter++

		data.Accounts[address] = &GenesisAccount{
			Address:       address,
			AccountNumber: fmt.Sprintf("%d", accountNumber),
			Sequence:      "0",
		}
	}

	// Parse balances for each denom in the header
	var coins []Coin
	for i := 1; i < len(header) && i < len(row); i++ {
		denom := header[i]
		amount := row[i]

		// Skip empty amounts
		if amount == "" || amount == "0" {
			continue
		}

		// Add coin
		coins = append(coins, Coin{
			Denom:  denom,
			Amount: amount,
		})
	}

	// Add to balances
	if len(coins) > 0 {
		if existing, ok := data.Balances[address]; ok {
			data.Balances[address] = append(existing, coins...)
		} else {
			data.Balances[address] = coins
		}
	}

	return nil
}

// Process supply.csv.
func processSupply(filePath string, data *GenesisData) error {
	// Open the CSV file
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open supply.csv: %w", err)
	}
	defer file.Close()

	// Create a new CSV reader
	reader := csv.NewReader(file)

	// Read the header
	header, err := reader.Read()
	if err != nil {
		return fmt.Errorf("failed to read header: %w", err)
	}

	// Expected header: denom,amount
	if len(header) < 2 || header[0] != "denom" || header[1] != "amount" {
		return errors.New("unexpected header format in supply.csv, expected: denom,amount")
	}

	// Process rows
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("error reading row: %w", err)
		}

		// Parse data
		denom := row[0]
		amount := row[1]

		// Add to supply
		data.Supply = append(data.Supply, Coin{
			Denom:  denom,
			Amount: amount,
		})
	}

	return nil
}

// Process pool_bals.csv and lp_bals.csv.
func processLPs(ipfsDir string) {
	// This is a simplified implementation that just reports files found
	// In a real-world scenario, you would need to process LP-related data

	// Check for pool_bals.csv
	poolFilePath := filepath.Join(ipfsDir, "pool_bals.csv")
	if _, err := os.Stat(poolFilePath); err == nil {
		fmt.Println("pool_bals.csv found, but specialized LP processing not implemented")
	}

	// Check for lp_bals.csv
	lpFilePath := filepath.Join(ipfsDir, "lp_bals.csv")
	if _, err := os.Stat(lpFilePath); err == nil {
		fmt.Println("lp_bals.csv found, but specialized LP processing not implemented")
	}

	// Check for total_lps.csv
	totalLPsFilePath := filepath.Join(ipfsDir, "total_lps.csv")
	if _, err := os.Stat(totalLPsFilePath); err == nil {
		fmt.Println("total_lps.csv found, but specialized LP processing not implemented")
	}
}

// Generate the final genesis.json file.
func generateGenesisJSON(data *GenesisData, chainID string) error {
	// Create auth genesis
	authGenesis := AuthGenesis{
		Params: AuthParams{
			MaxMemoCharacters:      "256",
			TxSigLimit:             "7",
			TxSizeCostPerByte:      "10",
			SigVerifyCostEd25519:   "590",
			SigVerifyCostSecp256k1: "1000",
		},
		Accounts: make([]GenesisAccount, 0, len(data.Accounts)),
	}

	// Convert accounts map to slice
	var accounts []GenesisAccount
	for _, account := range data.Accounts {
		accounts = append(accounts, *account)
	}

	// Sort accounts by account number
	sort.Slice(accounts, func(i, j int) bool {
		numI, _ := strconv.Atoi(accounts[i].AccountNumber)
		numJ, _ := strconv.Atoi(accounts[j].AccountNumber)
		return numI < numJ
	})

	// Add sorted accounts to auth genesis
	authGenesis.Accounts = accounts

	// Create bank genesis
	bankGenesis := BankGenesis{
		Params: BankParams{
			SendEnabled:        []SendEnabled{},
			DefaultSendEnabled: true,
		},
		Balances:      make([]Balance, 0, len(data.Balances)),
		Supply:        data.Supply,
		DenomMetadata: []Metadata{},
	}

	// Convert balances map to slice
	for address, coins := range data.Balances {
		bankGenesis.Balances = append(bankGenesis.Balances, Balance{
			Address: address,
			Coins:   coins,
		})
	}

	// Create app state
	appState := AppState{
		Auth:    authGenesis,
		Bank:    bankGenesis,
		Staking: map[string]any{},
		Genutil: map[string]any{},
	}

	// Create final genesis document
	genesisDoc := GenesisDoc{
		GenesisTime:   time.Now(),
		ChainID:       chainID,
		InitialHeight: "1",
		ConsensusParams: map[string]any{
			"block": map[string]any{
				"max_bytes": "22020096",
				"max_gas":   "-1",
			},
			"evidence": map[string]any{
				"max_age_num_blocks": "100000",
				"max_age_duration":   "172800000000000",
			},
			"validator": map[string]any{
				"pub_key_types": []string{"ed25519"},
			},
			"version": map[string]any{},
		},
		AppState: appState,
	}

	// Marshal the genesis document with pretty printing
	genesisBz, err := json.MarshalIndent(genesisDoc, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal genesis doc: %w", err)
	}

	// Write to file first
	tempFilePath := "genesis.json.temp"
	err = os.WriteFile(tempFilePath, genesisBz, 0o600)
	if err != nil {
		return fmt.Errorf("failed to write temporary genesis file: %w", err)
	}

	// Read the file back in
	fileContent, err := os.ReadFile(tempFilePath)
	if err != nil {
		return fmt.Errorf("failed to read temporary genesis file: %w", err)
	}

	// Convert to string for replacements
	jsonStr := string(fileContent)

	// Perform all prefix replacements
	jsonStr = strings.ReplaceAll(jsonStr, "\"unicorn1", "\"gadikian1")
	jsonStr = strings.ReplaceAll(jsonStr, "\"uwunicorn\"", "\"ugadikian\"")
	jsonStr = strings.ReplaceAll(jsonStr, "\"valuwunicorn\"", "\"valgadikian\"")
	jsonStr = strings.Replace(jsonStr, "\"chain_id\": \"unicorn-1\"", "\"chain_id\": \"gadikian-1\"", 1)

	// Log what we did
	fmt.Println("Converted bech32 prefixes from 'unicorn' to 'gadikian'")
	fmt.Println("Converted token denominations from 'uwunicorn' to 'ugadikian'")
	fmt.Println("Converted validator token denominations from 'valuwunicorn' to 'valgadikian'")
	fmt.Println("Set chain ID to 'gadikian-1'")

	// Write to final file
	err = os.WriteFile("genesis.json", []byte(jsonStr), 0o600)
	if err != nil {
		return fmt.Errorf("failed to write final genesis file: %w", err)
	}

	// Clean up temp file
	os.Remove(tempFilePath)

	return nil
}
