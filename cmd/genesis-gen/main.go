package main

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"os"
	"path/filepath"
	"time"

	sdkmath "cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/codec"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// Define a struct to hold the genesis data.
type GenesisData struct {
	Accounts    []authtypes.GenesisAccount
	Balances    []banktypes.Balance
	Supply      sdktypes.Coins
	StakingData stakingtypes.GenesisState
	// Add other modules as needed
}

// Define a struct to hold the final genesis document.
type GenesisDoc struct {
	GenesisTime     time.Time                  `json:"genesis_time"`
	ChainID         string                     `json:"chain_id"`
	InitialHeight   string                     `json:"initial_height"`
	ConsensusParams map[string]interface{}     `json:"consensus_params"`
	AppState        map[string]json.RawMessage `json:"app_state"`
}

func main() {
	// Check if IPFS directory is provided
	if len(os.Args) < 2 {
		fmt.Println("Usage: genesis-gen <ipfs-dir> [chain-id]")
		os.Exit(1)
	}

	ipfsDir := os.Args[1]
	chainID := "unicorn-1"
	if len(os.Args) > 2 {
		chainID = os.Args[2]
	}

	// Initialize genesis data
	genesisData := &GenesisData{
		Accounts: make([]authtypes.GenesisAccount, 0),
		Balances: make([]banktypes.Balance, 0),
		Supply:   sdktypes.Coins{},
	}

	// Process balances.csv if it exists
	fmt.Println("Processing balances.csv...")
	balancesPath := filepath.Join(ipfsDir, "balances.csv")
	if _, err := os.Stat(balancesPath); err == nil {
		err := processBalances(ipfsDir, genesisData)
		if err != nil {
			fmt.Printf("Error processing balances: %v\n", err)
			os.Exit(1)
		}
	} else {
		fmt.Println("balances.csv not found, skipping...")
	}

	// Process supply.csv
	fmt.Println("Processing supply.csv...")
	err := processSupply(ipfsDir, genesisData)
	if err != nil {
		fmt.Printf("Error processing supply: %v\n", err)
		os.Exit(1)
	}

	// Process kaway_bond.csv
	fmt.Println("Processing kaway_bond.csv...")
	err = processKawayBond(ipfsDir, genesisData)
	if err != nil {
		fmt.Printf("Error processing kaway_bond: %v\n", err)
		os.Exit(1)
	}

	// Process uwuval_bond.csv if it exists
	fmt.Println("Processing uwuval_bond.csv...")
	uwuvalPath := filepath.Join(ipfsDir, "uwuval_bond.csv")
	if _, err := os.Stat(uwuvalPath); err == nil {
		err := processUwuvalBond(ipfsDir, genesisData)
		if err != nil {
			fmt.Printf("Error processing uwuval_bond: %v\n", err)
			os.Exit(1)
		}
	} else {
		fmt.Println("uwuval_bond.csv not found, skipping...")
	}

	// Process pool_bals.csv and lp_bals.csv
	fmt.Println("Processing liquidity pool data...")
	err = processLPs(ipfsDir, genesisData)
	if err != nil {
		fmt.Printf("Error processing LPs: %v\n", err)
		os.Exit(1)
	}

	// Create final genesis.json
	fmt.Println("Generating genesis.json...")
	err = generateGenesisJSON(genesisData, chainID)
	if err != nil {
		fmt.Printf("Error generating genesis JSON: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Genesis file created successfully: genesis.json")
}

// Process balances.csv - expected format is more complex with multiple columns.
func processBalances(ipfsDir string, data *GenesisData) error {
	filePath := filepath.Join(ipfsDir, "balances.csv")

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
		return errors.New("unexpected header format in balances.csv, first column should be 'address'")
	}

	// Map to store balances by address
	balancesByAddress := make(map[string]sdktypes.Coins)

	// Map of account addresses
	accountAddresses := make(map[string]bool)

	// Process rows
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("error reading row: %v", err)
		}

		// Parse address and record that we've seen it
		address := row[0]
		accountAddresses[address] = true

		// Parse balances for each denom in the header
		coins := sdktypes.Coins{}
		for i := 1; i < len(header) && i < len(row); i++ {
			denom := header[i]
			amount := row[i]

			// Skip empty amounts
			if amount == "" || amount == "0" {
				continue
			}

			// Create a coin
			intAmount, ok := new(big.Int).SetString(amount, 10)
			if !ok {
				return fmt.Errorf("failed to parse amount %s for address %s", amount, address)
			}

			coin := sdktypes.NewCoin(denom, sdkmath.NewIntFromBigInt(intAmount))
			coins = coins.Add(coin)
		}

		// Add to balances map
		if existing, ok := balancesByAddress[address]; ok {
			balancesByAddress[address] = existing.Add(coins...)
		} else {
			balancesByAddress[address] = coins
		}
	}

	// Create accounts and balances
	for address, coins := range balancesByAddress {
		if coins.IsZero() {
			continue // Skip zero balances
		}

		accAddress, err := sdktypes.AccAddressFromBech32(address)
		if err != nil {
			return fmt.Errorf("invalid address %s: %v", address, err)
		}

		// Create base account
		baseAccount := authtypes.NewBaseAccount(accAddress, nil, 0, 0)

		// Add to accounts and balances
		data.Accounts = append(data.Accounts, baseAccount)
		data.Balances = append(data.Balances, banktypes.Balance{
			Address: address,
			Coins:   coins,
		})
	}

	// Create accounts for addresses without balances
	for address := range accountAddresses {
		if _, exists := balancesByAddress[address]; !exists || balancesByAddress[address].IsZero() {
			accAddress, err := sdktypes.AccAddressFromBech32(address)
			if err != nil {
				return fmt.Errorf("invalid address %s: %v", address, err)
			}

			// Create base account
			baseAccount := authtypes.NewBaseAccount(accAddress, nil, 0, 0)

			// Add to accounts
			data.Accounts = append(data.Accounts, baseAccount)
		}
	}

	return nil
}

// Process supply.csv.
func processSupply(ipfsDir string, data *GenesisData) error {
	filePath := filepath.Join(ipfsDir, "supply.csv")

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
		return errors.New("unexpected header format in supply.csv, expected: denom,amount")
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

		// Create a coin
		intAmount, ok := new(big.Int).SetString(amount, 10)
		if !ok {
			return fmt.Errorf("failed to parse amount %s for denom %s", amount, denom)
		}

		coin := sdktypes.NewCoin(denom, sdkmath.NewIntFromBigInt(intAmount))

		// Add to total supply
		data.Supply = data.Supply.Add(coin)
	}

	return nil
}

// Process kaway_bond.csv.
func processKawayBond(ipfsDir string, data *GenesisData) error {
	filePath := filepath.Join(ipfsDir, "kaway_bond.csv")

	// Open the CSV file
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open kaway_bond.csv: %v", err)
	}
	defer file.Close()

	// Create a new CSV reader
	reader := csv.NewReader(file)

	// Read the header
	header, err := reader.Read()
	if err != nil {
		return fmt.Errorf("failed to read header: %v", err)
	}

	// Expected header: address,uwu
	if len(header) < 2 || header[0] != "address" || header[1] != "uwu" {
		return errors.New("unexpected header format in kaway_bond.csv, expected: address,uwu")
	}

	// Map to track addresses that already have accounts
	existingAccounts := make(map[string]bool)
	for _, account := range data.Accounts {
		existingAccounts[account.GetAddress().String()] = true
	}

	// Map to track balances that already exist
	existingBalances := make(map[string]int)
	for i, balance := range data.Balances {
		existingBalances[balance.Address] = i
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

		// Create a coin (uwunicorn is the native token)
		intAmount, ok := new(big.Int).SetString(amount, 10)
		if !ok {
			return fmt.Errorf("failed to parse amount %s for address %s", amount, address)
		}

		coin := sdktypes.NewCoin("uwunicorn", sdkmath.NewIntFromBigInt(intAmount))

		// Add account if it doesn't exist
		if !existingAccounts[address] {
			accAddress, err := sdktypes.AccAddressFromBech32(address)
			if err != nil {
				return fmt.Errorf("invalid address %s: %v", address, err)
			}

			baseAccount := authtypes.NewBaseAccount(accAddress, nil, 0, 0)
			data.Accounts = append(data.Accounts, baseAccount)
			existingAccounts[address] = true
		}

		// Update balances
		if idx, exists := existingBalances[address]; exists {
			data.Balances[idx].Coins = data.Balances[idx].Coins.Add(coin)
		} else {
			data.Balances = append(data.Balances, banktypes.Balance{
				Address: address,
				Coins:   sdktypes.Coins{coin},
			})
			existingBalances[address] = len(data.Balances) - 1
		}
	}

	return nil
}

// Process uwuval_bond.csv - similar to kaway_bond.csv.
func processUwuvalBond(ipfsDir string, data *GenesisData) error {
	filePath := filepath.Join(ipfsDir, "uwuval_bond.csv")

	// Open the CSV file
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open uwuval_bond.csv: %v", err)
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
		return errors.New("unexpected header format in uwuval_bond.csv, expected: address,amount")
	}

	// Map to track addresses that already have accounts
	existingAccounts := make(map[string]bool)
	for _, account := range data.Accounts {
		existingAccounts[account.GetAddress().String()] = true
	}

	// Map to track balances that already exist
	existingBalances := make(map[string]int)
	for i, balance := range data.Balances {
		existingBalances[balance.Address] = i
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

		// Create a coin (valuwunicorn for validator tokens)
		intAmount, ok := new(big.Int).SetString(amount, 10)
		if !ok {
			return fmt.Errorf("failed to parse amount %s for address %s", amount, address)
		}

		coin := sdktypes.NewCoin("valuwunicorn", sdkmath.NewIntFromBigInt(intAmount))

		// Add account if it doesn't exist
		if !existingAccounts[address] {
			accAddress, err := sdktypes.AccAddressFromBech32(address)
			if err != nil {
				return fmt.Errorf("invalid address %s: %v", address, err)
			}

			baseAccount := authtypes.NewBaseAccount(accAddress, nil, 0, 0)
			data.Accounts = append(data.Accounts, baseAccount)
			existingAccounts[address] = true
		}

		// Update balances
		if idx, exists := existingBalances[address]; exists {
			data.Balances[idx].Coins = data.Balances[idx].Coins.Add(coin)
		} else {
			data.Balances = append(data.Balances, banktypes.Balance{
				Address: address,
				Coins:   sdktypes.Coins{coin},
			})
			existingBalances[address] = len(data.Balances) - 1
		}
	}

	return nil
}

// Process pool_bals.csv and lp_bals.csv.
func processLPs(ipfsDir string, data *GenesisData) error {
	// This is a simplified implementation
	// In a real-world scenario, you would need to process LP-related data
	// and update the appropriate state for modules like liquidity, x/gamm (for Osmosis-style pools), etc.
	// The exact implementation depends on your chain's modules

	// Add handling for pool_bals.csv
	poolFilePath := filepath.Join(ipfsDir, "pool_bals.csv")
	if _, err := os.Stat(poolFilePath); err == nil {
		fmt.Println("pool_bals.csv found, but specialized LP processing not implemented")
		// Implementation would go here
	}

	// Add handling for lp_bals.csv
	lpFilePath := filepath.Join(ipfsDir, "lp_bals.csv")
	if _, err := os.Stat(lpFilePath); err == nil {
		fmt.Println("lp_bals.csv found, but specialized LP processing not implemented")
		// Implementation would go here
	}

	// Add handling for total_lps.csv
	totalLPsFilePath := filepath.Join(ipfsDir, "total_lps.csv")
	if _, err := os.Stat(totalLPsFilePath); err == nil {
		fmt.Println("total_lps.csv found, but specialized LP processing not implemented")
		// Implementation would go here
	}

	return nil
}

// Generate the final genesis.json file.
func generateGenesisJSON(data *GenesisData, chainID string) error {
	// Create codec for encoding
	cdc := codec.NewLegacyAmino()
	authtypes.RegisterLegacyAminoCodec(cdc)
	banktypes.RegisterLegacyAminoCodec(cdc)
	stakingtypes.RegisterLegacyAminoCodec(cdc)

	// Initialize the app state
	appState := make(map[string]json.RawMessage)

	// Add auth state
	authGenesis := authtypes.NewGenesisState(authtypes.DefaultParams(), data.Accounts)
	authGenesisBz, err := cdc.MarshalJSON(authGenesis)
	if err != nil {
		return fmt.Errorf("failed to marshal auth genesis: %v", err)
	}
	appState["auth"] = authGenesisBz

	// Add bank state
	bankGenesis := banktypes.NewGenesisState(
		banktypes.DefaultParams(),
		data.Balances,
		data.Supply,
		[]banktypes.Metadata{},
		[]banktypes.SendEnabled{},
	)
	bankGenesisBz, err := cdc.MarshalJSON(bankGenesis)
	if err != nil {
		return fmt.Errorf("failed to marshal bank genesis: %v", err)
	}
	appState["bank"] = bankGenesisBz

	// Add staking state (simplified, you'll need to customize this)
	stakingGenesis := stakingtypes.DefaultGenesisState()
	stakingGenesisBz, err := cdc.MarshalJSON(stakingGenesis)
	if err != nil {
		return fmt.Errorf("failed to marshal staking genesis: %v", err)
	}
	appState["staking"] = stakingGenesisBz

	// Add genutil state (empty transactions for now)
	genutilGenesis := genutiltypes.NewGenesisState([]json.RawMessage{})
	genutilGenesisBz, err := cdc.MarshalJSON(genutilGenesis)
	if err != nil {
		return fmt.Errorf("failed to marshal genutil genesis: %v", err)
	}
	appState["genutil"] = genutilGenesisBz

	// Create the final genesis document
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

	// Marshal the genesis document
	genesisBz, err := json.MarshalIndent(genesisDoc, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal genesis doc: %v", err)
	}

	// Write to file
	err = os.WriteFile("genesis.json", genesisBz, 0o644)
	if err != nil {
		return fmt.Errorf("failed to write genesis file: %v", err)
	}

	return nil
}
