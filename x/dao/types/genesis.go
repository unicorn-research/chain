package types

// GenesisState defines the DAO module's genesis state.
type GenesisState struct {
	// DAOs lists the registered DAOs at genesis
	DAOs []DAO `json:"daos"`

	// Members lists the DAO members at genesis
	Members []Member `json:"members"`

	// Proposals lists the DAO proposals at genesis
	Proposals []Proposal `json:"proposals"`

	// Votes lists the votes at genesis
	Votes []Vote `json:"votes"`
}

// DefaultGenesis returns the default genesis state for the DAO module.
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		DAOs:      []DAO{},
		Members:   []Member{},
		Proposals: []Proposal{},
		Votes:     []Vote{},
	}
}

// Validate performs basic validation of genesis data returning an error for any
// failed validation criteria.
func (gs GenesisState) Validate() error {
	// Perform basic validation here, like checking for duplicate IDs, etc.
	return nil
}
