package types

import (
	"encoding/binary"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// DAOKeyPrefix is the prefix for DAO keys
	DAOKeyPrefix = "dao:"

	// MemberKeyPrefix is the prefix for DAO member keys
	MemberKeyPrefix = "member:"

	// ProposalKeyPrefix is the prefix for DAO proposal keys
	ProposalKeyPrefix = "proposal:"

	// VoteKeyPrefix is the prefix for vote keys
	VoteKeyPrefix = "vote:"
)

// GetDAOKey returns the store key to retrieve a DAO from its address
func GetDAOKey(daoAddr sdk.AccAddress) []byte {
	return append([]byte(DAOKeyPrefix), daoAddr.Bytes()...)
}

// GetMemberKey returns the store key to retrieve a member from the DAO and member address
func GetMemberKey(daoAddr, memberAddr sdk.AccAddress) []byte {
	key := append([]byte(MemberKeyPrefix), daoAddr.Bytes()...)
	return append(key, memberAddr.Bytes()...)
}

// GetProposalKey returns the store key to retrieve a proposal from its DAO address and ID
func GetProposalKey(daoAddr sdk.AccAddress, proposalID uint64) []byte {
	key := append([]byte(ProposalKeyPrefix), daoAddr.Bytes()...)
	idBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(idBytes, proposalID)
	return append(key, idBytes...)
}

// GetVoteKey returns the store key to retrieve a vote from its voter address and proposal ID
func GetVoteKey(voterAddr sdk.AccAddress, proposalID uint64) []byte {
	key := append([]byte(VoteKeyPrefix), voterAddr.Bytes()...)
	idBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(idBytes, proposalID)
	return append(key, idBytes...)
}
