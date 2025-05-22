package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// Define types for our genesis file
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
	GenesisTime     time.Time              `json:"genesis_time"`
	ChainID         string                 `json:"chain_id"`
	InitialHeight   string                 `json:"initial_height"`
	ConsensusParams map[string]interface{} `json:"consensus_params"`
	AppState        AppState               `json:"app_state"`
}

type AppState struct {
	Auth    AuthGenesis            `json:"auth"`
	Bank    BankGenesis            `json:"bank"`
	Staking map[string]interface{} `json:"staking"`
	Genutil map[string]interface{} `json:"genutil"`
	// Other modules can be added as needed
}

// Holds our internal processing data
type GenesisData struct {
	Accounts map[string]*GenesisAccount
	Balances map[string][]Coin
	Supply   []Coin
}

func main() {
	// Check if IPFS directory is provided
	if len(os.Args) < 2 {
		fmt.Println("Usage: genesis-tool <ipfs-dir> [chain-id]")
		os.Exit(1)
	}

	ipfsDir := os.Args[1]
	chainID := "unicorn-1"
	if len(os.Args) > 2 {
		chainID = os.Args[2]
	}

	// Initialize genesis data
	genesisData := &GenesisData{
		Accounts: make(map[string]*GenesisAccount),
		Balances: make(map[string][]Coin),
		Supply:   make([]Coin, 0),
	}

	// Process balances.csv if it exists
	fmt.Println("Processing balances.csv...")
	balancesPath := filepath.Join(ipfsDir, "balances.csv")
	if _, err := os.Stat(balancesPath); err == nil {
		err := processBalances(balancesPath, genesisData)
		if err != nil {
			fmt.Printf("Error processing balances: %v\n", err)
			os.Exit(1)
		}
	} else {
		fmt.Println("balances.csv not found, skipping...")
	}

	// Process supply.csv
	fmt.Println("Processing supply.csv...")
	supplyPath := filepath.Join(ipfsDir, "supply.csv")
	err := processSupply(supplyPath, genesisData)
	if err != nil {
		fmt.Printf("Error processing supply: %v\n", err)
		os.Exit(1)
	}

	// Process kaway_bond.csv
	fmt.Println("Processing kaway_bond.csv...")
	kawayPath := filepath.Join(ipfsDir, "kaway_bond.csv")
	err = processBonds(kawayPath, "uwunicorn", genesisData)
	if err != nil {
		fmt.Printf("Error processing kaway_bond: %v\n", err)
		os.Exit(1)
	}

	// Process uwuval_bond.csv if it exists
	fmt.Println("Processing uwuval_bond.csv...")
	uwuvalPath := filepath.Join(ipfsDir, "uwuval_bond.csv")
	if _, err := os.Stat(uwuvalPath); err == nil {
		err := processBonds(uwuvalPath, "valuwunicorn", genesisData)
		if err != nil {
			fmt.Printf("Error processing uwuval_bond: %v\n", err)
			os.Exit(1)
		}
	} else {
		fmt.Println("uwuval_bond.csv not found, skipping...")
	}

	// Process LP files
	fmt.Println("Processing liquidity pool data...")
	processLPs(ipfsDir, genesisData)

	// Create final genesis.json
	fmt.Println("Generating genesis.json...")
	err = generateGenesisJSON(genesisData, chainID)
	if err != nil {
		fmt.Printf("Error generating genesis JSON: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Genesis file created successfully: genesis.json")
}

// Process bonds from a CSV file (kaway_bond.csv or uwuval_bond.csv)
func processBonds(filePath, denom string, data *GenesisData) error {
	// Check if file exists
	if _, err := os.Stat(filePath); err != nil {
		fmt.Printf("%s not found, skipping...\n", filepath.Base(filePath))
		return nil
	}

	// Open the CSV file
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open %s: %v", filepath.Base(filePath), err)
	}
	defer file.Close()

	// Create a new CSV reader
	reader := csv.NewReader(file)

	// Read the header
	header, err := reader.Read()
	if err != nil {
		return fmt.Errorf("failed to read header: %v", err)
	}

	// Expected header: address,uwu or similar
	if len(header) < 2 || header[0] != "address" {
		return fmt.Errorf("unexpected header format in %s, expected: address,amount", filepath.Base(filePath))
	}

	// Process rows
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("error reading row: %v", err)
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
			data.Accounts[address] = &GenesisAccount{
				Address:       address,
				AccountNumber: "0",
				Sequence:      "0",
			}
		}

		// Add coin to balances
		if _, exists := data.Balances[address]; exists {
			data.Balances[address] = append(data.Balances[address], coin)
		} else {
			data.Balances[address] = []Coin{coin}
		}
	}

	return nil
}

// Process balances.csv - expected to be complex with many columns
func processBalances(filePath string, data *GenesisData) error {
	// Open the CSV file
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open balances.csv: %v", err)
	}
	defer file.Close()

	// Create a new CSV reader
	reader := csv.NewReader(file)

	// Read the header to determine columns
	header, err := reader.Read()
	if err != nil {
		return fmt.Errorf("failed to read header: %v", err)
	}

	// First column should be 'address'
	if header[0] != "address" {
		return fmt.Errorf("unexpected header format in balances.csv, first column should be 'address'")
	}

	// Process rows
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("error reading row: %v", err)
		}

		// Parse address
		address := row[0]

		// Ensure account exists
		if _, exists := data.Accounts[address]; !exists {
			data.Accounts[address] = &GenesisAccount{
				Address:       address,
				AccountNumber: "0",
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
	}

	return nil
}

// Process supply.csv
func processSupply(filePath string, data *GenesisData) error {
	// Open the CSV file
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open supply.csv: %v", err)
	}
	defer file.Close()

	// Create a new CSV reader
	reader := csv.NewReader(file)

	// Read the header
	header, err := reader.Read()
	if err != nil {
		return fmt.Errorf("failed to read header: %v", err)
	}

	// Expected header: denom,amount
	if len(header) < 2 || header[0] != "denom" || header[1] != "amount" {
		return fmt.Errorf("unexpected header format in supply.csv, expected: denom,amount")
	}

	// Process rows
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("error reading row: %v", err)
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

// Process pool_bals.csv and lp_bals.csv
func processLPs(ipfsDir string, data *GenesisData) error {
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

	return nil
}

// Generate the final genesis.json file
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
	for _, account := range data.Accounts {
		authGenesis.Accounts = append(authGenesis.Accounts, *account)
	}

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
		Staking: map[string]interface{}{},
		Genutil: map[string]interface{}{},
	}

	// Create final genesis document
	genesisDoc := GenesisDoc{
		GenesisTime:   time.Now(),
		ChainID:       chainID,
		InitialHeight: "1",
		ConsensusParams: map[string]interface{}{
			"block": map[string]interface{}{
				"max_bytes": "22020096",
				"max_gas":   "-1",
			},
			"evidence": map[string]interface{}{
				"max_age_num_blocks": "100000",
				"max_age_duration":   "172800000000000",
			},
			"validator": map[string]interface{}{
				"pub_key_types": []string{"ed25519"},
			},
			"version": map[string]interface{}{},
		},
		AppState: appState,
	}

	// Marshal the genesis document with pretty printing
	genesisBz, err := json.MarshalIndent(genesisDoc, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal genesis doc: %v", err)
	}

	// Write to file
	err = os.WriteFile("genesis.json", genesisBz, 0644)
	if err != nil {
		return fmt.Errorf("failed to write genesis file: %v", err)
	}

	return nil
}
