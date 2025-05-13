package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const (
	// TypeMsgCreateDAO is the type for MsgCreateDAO
	TypeMsgCreateDAO = "create_dao"

	// TypeMsgJoinDAO is the type for MsgJoinDAO
	TypeMsgJoinDAO = "join_dao"

	// TypeMsgSubmitProposal is the type for MsgSubmitProposal
	TypeMsgSubmitProposal = "submit_dao_proposal"

	// TypeMsgVote is the type for MsgVote
	TypeMsgVote = "dao_vote"

	// TypeMsgWithdraw is the type for MsgWithdraw
	TypeMsgWithdraw = "withdraw_from_dao"

	// TypeMsgAmendArticles is the type for MsgAmendArticles
	TypeMsgAmendArticles = "amend_dao_articles"

	// TypeMsgDissolveDAO is the type for MsgDissolveDAO
	TypeMsgDissolveDAO = "dissolve_dao"
)

// MsgCreateDAO defines the message to create a new DAO
type MsgCreateDAO struct {
	// Creator is the address of the person creating the DAO
	Creator sdk.AccAddress `json:"creator"`

	// Name of the DAO (must include DAO, LAO or DAO LLC)
	Name string `json:"name"`

	// Description of the DAO
	Description string `json:"description"`

	// PublicIdentifier is the required publicly available identifier
	PublicIdentifier string `json:"public_identifier"`

	// Articles defines the articles of organization
	Articles ArticlesOfOrganization `json:"articles"`

	// OperatingAgreement contains the optional operating rules
	OperatingAgreement string `json:"operating_agreement,omitempty"`

	// InitialMembers are the founding members of the DAO
	InitialMembers []sdk.AccAddress `json:"initial_members"`

	// ExpirationDate is the optional period fixed for the duration
	ExpirationDate *int64 `json:"expiration_date,omitempty"`
}

// ValidateBasic performs stateless validation of MsgCreateDAO
func (msg MsgCreateDAO) ValidateBasic() error {
	// Validate Creator is not empty
	if msg.Creator.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "creator address cannot be empty")
	}

	// Validate DAO name has proper suffix per Wyoming spec
	if !hasDaoSuffix(msg.Name) {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest,
			"DAO name must include 'DAO', 'LAO', or 'DAO LLC' suffix per Wyoming specs")
	}

	// Validate PublicIdentifier is not empty
	if msg.PublicIdentifier == "" {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "public identifier cannot be empty")
	}

	// Validate Articles
	if err := validateArticles(msg.Articles); err != nil {
		return err
	}

	// Must have at least one initial member (the creator)
	if len(msg.InitialMembers) == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "DAO must have at least one member")
	}

	return nil
}

// GetSigners returns the signers of the message
func (msg MsgCreateDAO) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Creator}
}

// hasDaoSuffix validates that the name has proper DAO suffix per Wyoming statute
func hasDaoSuffix(name string) bool {
	// Check for DAO, LAO, or DAO LLC suffix
	return sdk.StrContains(name, "DAO") ||
		sdk.StrContains(name, "LAO") ||
		sdk.StrContains(name, "DAO LLC")
}

// validateArticles validates that the ArticlesOfOrganization has all required fields
func validateArticles(articles ArticlesOfOrganization) error {
	// Wyoming DAO spec requires these statements in the articles
	if articles.StatementOfDAO == "" {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "statement of DAO is required")
	}
	if articles.ManagementStatement == "" {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "management statement is required")
	}
	// Validate other required article fields

	return nil
}

// MsgJoinDAO defines the message to join an existing DAO
type MsgJoinDAO struct {
	// DAO address to join
	DAOAddress sdk.AccAddress `json:"dao_address"`

	// Member address joining the DAO
	Member sdk.AccAddress `json:"member"`

	// Contribution is the optional assets being contributed
	Contribution sdk.Coins `json:"contribution,omitempty"`
}

// ValidateBasic performs stateless validation of MsgJoinDAO
func (msg MsgJoinDAO) ValidateBasic() error {
	if msg.DAOAddress.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "DAO address cannot be empty")
	}
	if msg.Member.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "member address cannot be empty")
	}

	return nil
}

// GetSigners returns the signers of the message
func (msg MsgJoinDAO) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Member}
}

// MsgWithdraw defines the message to withdraw from a DAO
type MsgWithdraw struct {
	// DAO address to withdraw from
	DAOAddress sdk.AccAddress `json:"dao_address"`

	// Member address withdrawing
	Member sdk.AccAddress `json:"member"`
}

// ValidateBasic performs stateless validation of MsgWithdraw
func (msg MsgWithdraw) ValidateBasic() error {
	if msg.DAOAddress.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "DAO address cannot be empty")
	}
	if msg.Member.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "member address cannot be empty")
	}

	return nil
}

// GetSigners returns the signers of the message
func (msg MsgWithdraw) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Member}
}
