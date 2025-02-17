# Solana Pump Package

A Go package for interacting with Pump fun contract, enabling token buying and selling operations.
This class was used in my personal pump fun coin sniper.

## Features

- Support for Buy and Sell operations
- Automatic slippage calculation
- Associated Token Account handling

## Prerequisites
- Go 1.19 or higher
- Dependencies:
  - github.com/gagliardetto/solana-go
  - Standard Go packages

## Installation
```bash
go get github.com/frigge/solana-pump
```

## Usage

### Initialization

```go
import (
    "github.com/frigge/pump"
    solana "github.com/gagliardetto/solana-go"
)

coinToBuy := Coin{}
coinToBuy.SetBuyAmount(0.02) //In Sol
coinToBuy.SetSlippage(0.6)  //0.0 - 1.0
coinToBuy.SetMintAddress(MintAddress) //Coin mint address
```

### Buying Tokens

```go
// Setup RPC client
client := rpc.New(YourRpcUrl)

// Execute purchase
signature, err := coin.Buy(client, YourPrivateKey)
if err != nil {
    log.Fatalf("Error during purchase: %v", err)
}
fmt.Printf("Transaction completed: %s\n", signature)
```

### Selling Tokens

```go
// Execute sale
signature, err := coin.Sell(client, ownerPrivateKey)
if err != nil {
    log.Fatalf("Error during sale: %v", err)
}
fmt.Printf("Transaction completed: %s\n", signature)
```

## BondingCurve Structure

```go
type BondingCurve struct {
    VirtualTokenReserves uint64
    VirtualSolReserves   uint64
    RealTokenReserves    uint64
    RealSolReserves      uint64
    TokenTotalSupply     uint64
    Complete             bool
}
```

## Core Functions

### Coin Methods
- `SetBuyAmount(amount float64)`: Set the amount of SOL to use
- `SetSlippage(slippage float64)`: Set transaction slippage
- `SetTokenAddress(address solana.PublicKey)`: Set token address
- `SetCurveAddress(address solana.PublicKey)`: Set bonding curve address
- `Buy(client rpc.Client, owner solana.PrivateKey)`: Execute purchase
- `Sell(client rpc.Client, owner solana.PrivateKey)`: Execute sale

### Utility Functions
- `getBondingCurveInfos`: Retrieve bonding curve information
- `getAssociatedTokenAddress`: Calculate Associated Token Account address
- `decodeBondingData`: Decode bonding curve data

## Advanced Usage

### Transaction Options

The package is configured with optimized transaction settings for maximum speed:

- Uses a compute budget of 72,000 units
- Sets compute unit price to 517,000 micro-lamports
- Skips preflight checks to minimize transaction time
- Sets commitment to Finalized for final confirmation

## Contributing

Contributions and pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change.

Please make sure to update tests as appropriate.

## License

MIT

## Disclaimer

This software is provided "as is", without warranty of any kind. Use at your own risk. The developers assume no liability for any losses or damages resulting from the use of this package.

## Support

For bugs, feature requests, or support questions:
- Open an issue in the GitHub repository
- Contact the maintainers