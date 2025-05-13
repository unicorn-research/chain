package keeper

import (
	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/unicorn-reseaerch/chain/x/dao/types"
)

// Keeper manages DAO state
type Keeper struct {
	storeKey storetypes.StoreKey
	cdc      codec.BinaryCodec

	// Add dependencies for bank operations, auth, and other modules
	bankKeeper types.BankKeeper

	// Logger for the module
	logger log.Logger
}

// NewKeeper creates a new DAO Keeper instance
func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey storetypes.StoreKey,
	bankKeeper types.BankKeeper,
	logger log.Logger,
) Keeper {
	return Keeper{
		storeKey:   storeKey,
		cdc:        cdc,
		bankKeeper: bankKeeper,
		logger:     logger,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return k.logger.With("module", "x/"+types.ModuleName)
}

// InitGenesis initializes the DAO module's state from a given genesis state.
func (k Keeper) InitGenesis(ctx sdk.Context, genState types.GenesisState) {
	// Save all DAOs to state
	for _, dao := range genState.DAOs {
		k.SetDAO(ctx, dao)
	}

	// Save all DAO members
	for _, member := range genState.Members {
		k.SetMember(ctx, member)
	}

	// Save all proposals
	for _, proposal := range genState.Proposals {
		k.SetProposal(ctx, proposal)
	}

	// Save all votes
	for _, vote := range genState.Votes {
		k.SetVote(ctx, vote)
	}
}

// ExportGenesis returns the DAO module's genesis state.
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	return &types.GenesisState{
		DAOs:      k.GetAllDAOs(ctx),
		Members:   k.GetAllMembers(ctx),
		Proposals: k.GetAllProposals(ctx),
		Votes:     k.GetAllVotes(ctx),
	}
}

// SetDAO saves a DAO to the store
func (k Keeper) SetDAO(ctx sdk.Context, dao types.DAO) {
	store := ctx.KVStore(k.storeKey)
	daoKey := types.GetDAOKey(dao.Address)
	daoBytes := k.cdc.MustMarshal(&dao)
	store.Set(daoKey, daoBytes)
}

// GetDAO retrieves a DAO by address
func (k Keeper) GetDAO(ctx sdk.Context, daoAddr sdk.AccAddress) (types.DAO, bool) {
	store := ctx.KVStore(k.storeKey)
	daoKey := types.GetDAOKey(daoAddr)

	bz := store.Get(daoKey)
	if bz == nil {
		return types.DAO{}, false
	}

	var dao types.DAO
	k.cdc.MustUnmarshal(bz, &dao)
	return dao, true
}

// GetAllDAOs retrieves all DAOs
func (k Keeper) GetAllDAOs(ctx sdk.Context) []types.DAO {
	var daos []types.DAO
	store := ctx.KVStore(k.storeKey)

	iterator := sdk.KVStorePrefixIterator(store, []byte(types.DAOKeyPrefix))
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var dao types.DAO
		k.cdc.MustUnmarshal(iterator.Value(), &dao)
		daos = append(daos, dao)
	}

	return daos
}

// SetMember saves a Member to the store
func (k Keeper) SetMember(ctx sdk.Context, member types.Member) {
	store := ctx.KVStore(k.storeKey)
	memberKey := types.GetMemberKey(member.DAOAddress, member.Address)
	memberBytes := k.cdc.MustMarshal(&member)
	store.Set(memberKey, memberBytes)
}

// GetMember retrieves a Member by DAO address and member address
func (k Keeper) GetMember(ctx sdk.Context, daoAddr, memberAddr sdk.AccAddress) (types.Member, bool) {
	store := ctx.KVStore(k.storeKey)
	memberKey := types.GetMemberKey(daoAddr, memberAddr)

	bz := store.Get(memberKey)
	if bz == nil {
		return types.Member{}, false
	}

	var member types.Member
	k.cdc.MustUnmarshal(bz, &member)
	return member, true
}

// GetAllMembers retrieves all Members
func (k Keeper) GetAllMembers(ctx sdk.Context) []types.Member {
	var members []types.Member
	store := ctx.KVStore(k.storeKey)

	iterator := sdk.KVStorePrefixIterator(store, []byte(types.MemberKeyPrefix))
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var member types.Member
		k.cdc.MustUnmarshal(iterator.Value(), &member)
		members = append(members, member)
	}

	return members
}

// SetProposal saves a Proposal to the store
func (k Keeper) SetProposal(ctx sdk.Context, proposal types.Proposal) {
	store := ctx.KVStore(k.storeKey)
	proposalKey := types.GetProposalKey(proposal.DAOAddress, proposal.ID)
	proposalBytes := k.cdc.MustMarshal(&proposal)
	store.Set(proposalKey, proposalBytes)
}

// GetProposal retrieves a Proposal by DAO address and proposal ID
func (k Keeper) GetProposal(ctx sdk.Context, daoAddr sdk.AccAddress, proposalID uint64) (types.Proposal, bool) {
	store := ctx.KVStore(k.storeKey)
	proposalKey := types.GetProposalKey(daoAddr, proposalID)

	bz := store.Get(proposalKey)
	if bz == nil {
		return types.Proposal{}, false
	}

	var proposal types.Proposal
	k.cdc.MustUnmarshal(bz, &proposal)
	return proposal, true
}

// GetAllProposals retrieves all Proposals
func (k Keeper) GetAllProposals(ctx sdk.Context) []types.Proposal {
	var proposals []types.Proposal
	store := ctx.KVStore(k.storeKey)

	iterator := sdk.KVStorePrefixIterator(store, []byte(types.ProposalKeyPrefix))
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var proposal types.Proposal
		k.cdc.MustUnmarshal(iterator.Value(), &proposal)
		proposals = append(proposals, proposal)
	}

	return proposals
}

// SetVote saves a Vote to the store
func (k Keeper) SetVote(ctx sdk.Context, vote types.Vote) {
	store := ctx.KVStore(k.storeKey)
	voteKey := types.GetVoteKey(vote.Voter, vote.ProposalID)
	voteBytes := k.cdc.MustMarshal(&vote)
	store.Set(voteKey, voteBytes)
}

// GetVote retrieves a Vote by voter address and proposal ID
func (k Keeper) GetVote(ctx sdk.Context, voterAddr sdk.AccAddress, proposalID uint64) (types.Vote, bool) {
	store := ctx.KVStore(k.storeKey)
	voteKey := types.GetVoteKey(voterAddr, proposalID)

	bz := store.Get(voteKey)
	if bz == nil {
		return types.Vote{}, false
	}

	var vote types.Vote
	k.cdc.MustUnmarshal(bz, &vote)
	return vote, true
}

// GetAllVotes retrieves all Votes
func (k Keeper) GetAllVotes(ctx sdk.Context) []types.Vote {
	var votes []types.Vote
	store := ctx.KVStore(k.storeKey)

	iterator := sdk.KVStorePrefixIterator(store, []byte(types.VoteKeyPrefix))
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var vote types.Vote
		k.cdc.MustUnmarshal(iterator.Value(), &vote)
		votes = append(votes, vote)
	}

	return votes
}

// CheckForInactiveDaos checks for DAOs that haven't had any actions in a year
// and marks them for dissolution as per Wyoming spec 17-31-114(a)(iv)
func (k Keeper) CheckForInactiveDaos(ctx sdk.Context) {
	currentTime := ctx.BlockTime()
	oneYearAgo := currentTime.AddDate(-1, 0, 0)

	// Iterate through all DAOs
	for _, dao := range k.GetAllDAOs(ctx) {
		// If the DAO hasn't had any action in a year, mark it for dissolution
		if dao.LastActionDate.Before(oneYearAgo) && dao.Status != "dissolved" {
			dao.Status = "dissolved"
			k.SetDAO(ctx, dao)

			// Emit an event or log for the dissolution
			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					"dao_dissolved_inactivity",
					sdk.NewAttribute("dao_address", dao.Address.String()),
					sdk.NewAttribute("last_action_date", dao.LastActionDate.String()),
				),
			)
		}
	}
}
