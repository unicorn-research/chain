# Genesis Tool for Unicorn Chain

This tool generates a `genesis.json` file for Unicorn Chain from CSV files containing account balances, token supply, and other chain state.

## Build

```bash
go build -o genesis-tool
```

## Usage

```bash
./genesis-tool <ipfs-dir> [chain-id]
```

- `<ipfs-dir>`: Path to the directory containing the CSV files (e.g., `QmNyt5bh6KRgPukeH2XScdRnycn4pxHVyAdMKgrHVMktGX`)
- `[chain-id]`: (Optional) Chain ID to use in the genesis file (default: "unicorn-1")

## Expected CSV Files

The tool looks for the following CSV files in the IPFS directory:

### Required Files:

- **supply.csv**: Contains the total supply of each token
  - Format: `denom,amount`

### Optional Files:

- **balances.csv**: Contains account balances for various tokens
  - Format: The first column must be `address`, followed by denom columns
  
- **kaway_bond.csv**: Contains account bonded balances of uwunicorn tokens
  - Format: `address,uwu`
  
- **uwuval_bond.csv**: Contains validator bonded balances
  - Format: `address,uwu`
  
- **pool_bals.csv**: Contains liquidity pool balances
  - Format: `denom,uwu,meme`
  
- **lp_bals.csv**: Contains liquidity provider balances
  - First column is `address`, followed by LP token balances
  
- **total_lps.csv**: Contains total LP token shares
  - Format: `denom,shares`

## Output

The tool will generate a `genesis.json` file in the current directory with the following modules configured:

- `auth`: Account information
- `bank`: Token balances and supply
- `staking`: Empty staking parameters
- `genutil`: Empty genesis transactions

This genesis file can be used to start a new chain with the specified token distribution.

## Notes

- The LP processing is currently a placeholder and doesn't implement specialized handling for liquidity pools.
- This implementation uses a simplified approach with direct JSON structures rather than the SDK types.
- For production use, additional modules may need to be configured based on your chain's requirements. 