package types

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// ModuleName defines the module name
	ModuleName = "dao"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName
)

// DAO represents a Decentralized Autonomous Organization as per Wyoming spec
type DAO struct {
	// Name is required to include DAO, LAO or DAO LLC
	Name string `json:"name"`

	// Description of the DAO
	Description string `json:"description"`

	// Address is the account address of the DAO
	Address sdk.AccAddress `json:"address"`

	// ArticlesOfOrganization contains the required statements as per Wyoming spec
	ArticlesOfOrganization ArticlesOfOrganization `json:"articles_of_organization"`

	// OperatingAgreement contains optional operating rules
	OperatingAgreement string `json:"operating_agreement,omitempty"`

	// PublicIdentifier is the required publicly available identifier
	PublicIdentifier string `json:"public_identifier"`

	// CreatedAt is when the DAO was registered
	CreatedAt time.Time `json:"created_at"`

	// Status of the DAO - active, dissolved, etc.
	Status string `json:"status"`

	// ExpirationDate is optional period fixed for the duration
	ExpirationDate *time.Time `json:"expiration_date,omitempty"`

	// LastActionDate is when the DAO last took an action (to track the 1-year inactivity rule)
	LastActionDate time.Time `json:"last_action_date"`
}

// ArticlesOfOrganization contains the required statements from Wyoming spec
type ArticlesOfOrganization struct {
	// StatementOfDAO confirms this is a DAO
	StatementOfDAO string `json:"statement_of_dao"`

	// ManagementStatement explains how the DAO is managed
	ManagementStatement string `json:"management_statement"`

	// RightsAndDutiesOfMembers defines member rights and duties
	RightsAndDutiesOfMembers string `json:"rights_and_duties_of_members"`

	// TransferabilityOfInterests defines if/how membership interests can be transferred
	TransferabilityOfInterests string `json:"transferability_of_interests"`

	// WithdrawalRules defines how members can withdraw
	WithdrawalRules string `json:"withdrawal_rules"`

	// DistributionRules defines how assets are distributed to members
	DistributionRules string `json:"distribution_rules"`

	// AmendmentProcedures defines how articles can be amended
	AmendmentProcedures string `json:"amendment_procedures"`

	// DisputeResolutionRules defines how disputes are handled
	DisputeResolutionRules string `json:"dispute_resolution_rules"`
}

// Member represents a member of a DAO
type Member struct {
	// Address of the member
	Address sdk.AccAddress `json:"address"`

	// DAOAddress the member belongs to
	DAOAddress sdk.AccAddress `json:"dao_address"`

	// MembershipInterest represents the voting weight/ownership interest
	MembershipInterest string `json:"membership_interest"`

	// Contribution is the amount of assets contributed (for calculating interest)
	Contribution sdk.Coins `json:"contribution"`

	// JoinedAt is when the member joined
	JoinedAt time.Time `json:"joined_at"`
}

// Proposal represents a governance proposal in a DAO
type Proposal struct {
	// ID of the proposal
	ID uint64 `json:"id"`

	// DAOAddress the proposal belongs to
	DAOAddress sdk.AccAddress `json:"dao_address"`

	// Title of the proposal
	Title string `json:"title"`

	// Description of what the proposal does
	Description string `json:"description"`

	// ProposalType categorizes what the proposal does
	ProposalType string `json:"proposal_type"`

	// Proposer address
	Proposer sdk.AccAddress `json:"proposer"`

	// Status of the proposal
	Status string `json:"status"`

	// VotingEndTime when voting ends
	VotingEndTime time.Time `json:"voting_end_time"`

	// YesVotes is the total yes votes
	YesVotes string `json:"yes_votes"`

	// NoVotes is the total no votes
	NoVotes string `json:"no_votes"`

	// Executed indicates if the proposal was executed
	Executed bool `json:"executed"`

	// ExecutionData contains data for execution based on proposal type
	ExecutionData string `json:"execution_data,omitempty"`
}

// Vote represents a member's vote on a proposal
type Vote struct {
	// Voter is the address of the voter
	Voter sdk.AccAddress `json:"voter"`

	// ProposalID is the ID of the proposal
	ProposalID uint64 `json:"proposal_id"`

	// VoteOption is yes or no
	VoteOption string `json:"vote_option"`

	// VotingPower is the voting weight
	VotingPower string `json:"voting_power"`
}
