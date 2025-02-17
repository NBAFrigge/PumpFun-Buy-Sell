package pump

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"math"

	solana "github.com/gagliardetto/solana-go"
	associatedtokenaccount "github.com/gagliardetto/solana-go/programs/associated-token-account"
	computeBudget "github.com/gagliardetto/solana-go/programs/compute-budget"
	"github.com/gagliardetto/solana-go/rpc"
)

type BondingCurve struct {
	VirtualTokenReserves uint64
	VirtualSolReserves   uint64
	RealTokenReserves    uint64
	RealSolReserves      uint64
	TokenTotalSupply     uint64
	Complete             bool
}

type Coin struct {
	buyAmount              float64
	slippage               float64
	mintAddress            solana.PublicKey
	bondingCurveAddress    solana.PublicKey
	associatedBondingCurve solana.PublicKey
	coinBought             float64
	tokenAddress           solana.PublicKey
}

func decodeBondingData(bondingCurveData []byte) (*BondingCurve, error) {
	data := bondingCurveData[8:]

	bc := BondingCurve{}

	buf := bytes.NewReader(data)

	err := binary.Read(buf, binary.LittleEndian, &bc.VirtualTokenReserves)
	if err != nil {
		return nil, fmt.Errorf("error reading VirtualTokenReserves: %v", err)
	}

	err = binary.Read(buf, binary.LittleEndian, &bc.VirtualSolReserves)
	if err != nil {
		return nil, fmt.Errorf("error reading VirtualSolReserves: %v", err)
	}

	err = binary.Read(buf, binary.LittleEndian, &bc.RealTokenReserves)
	if err != nil {
		return nil, fmt.Errorf("error reading RealTokenReserves: %v", err)
	}

	err = binary.Read(buf, binary.LittleEndian, &bc.RealSolReserves)
	if err != nil {
		return nil, fmt.Errorf("error reading RealSolReserves: %v", err)
	}

	err = binary.Read(buf, binary.LittleEndian, &bc.TokenTotalSupply)
	if err != nil {
		return nil, fmt.Errorf("error reading TokenTotalSupply: %v", err)
	}

	var complete byte
	err = binary.Read(buf, binary.LittleEndian, &complete)
	if err != nil {
		return nil, fmt.Errorf("error reading Complete: %v", err)
	}
	bc.Complete = complete != 0

	return &bc, nil
}

func getBondingCurveInfos(client *rpc.Client, bondingCurve solana.PublicKey) (*BondingCurve, error) {
	mint, err := client.GetAccountInfoWithOpts(context.Background(), bondingCurve, &rpc.GetAccountInfoOpts{Commitment: rpc.CommitmentConfirmed})

	if err != nil {
		return nil, err
	}

	DataBytes := mint.Value.Data.GetBinary()
	Data, err := decodeBondingData(DataBytes)

	if err != nil {
		return nil, err
	}
	return Data, nil
}

func getAssociatedTokenAddress(walletAddress, mintAddress solana.PublicKey) (solana.PublicKey, error) {

	TokenAddress, _, err := solana.FindAssociatedTokenAddress(walletAddress, mintAddress)
	if err != nil {
		return solana.PublicKey{}, err
	}

	return TokenAddress, nil
}

func (c *Coin) GetTokenMint() solana.PublicKey {
	return c.mintAddress
}

func (c *Coin) SetBuyAmount(buyAmount float64) {
	c.buyAmount = buyAmount
}

func (c *Coin) SetSlippage(slippage float64) {
	c.slippage = slippage
}

func (c *Coin) SetTokenAddress(tokenAddress solana.PublicKey) {
	c.mintAddress = tokenAddress
}

func (c *Coin) SetCurveAddress(bondingCurveAddress solana.PublicKey) {
	c.bondingCurveAddress = bondingCurveAddress
}

func (c *Coin) SetAssociatedCurveAddress(associatedBondingCurve solana.PublicKey) {
	c.associatedBondingCurve = associatedBondingCurve
}

func (c *Coin) SetMintAddress(mintAddress solana.PublicKey) {
	c.mintAddress = mintAddress
}

func (c *Coin) Buy(client rpc.Client, Owner solana.PrivateKey) (solana.Signature, error) {
	TokenAddress, err := getAssociatedTokenAddress(Owner.PublicKey(), c.mintAddress)
	if err != nil {
		return solana.Signature{}, fmt.Errorf("error getting associated token address: %v", err)
	}

	c.tokenAddress = TokenAddress

	BondingCurveData, err := getBondingCurveInfos(&client, c.bondingCurveAddress)
	if err != nil {
		return solana.Signature{}, fmt.Errorf("error getting bonding curve data: %v", err)
	}

	solTokenPrice := float64(float64(BondingCurveData.VirtualSolReserves/solana.LAMPORTS_PER_SOL)) / float64(BondingCurveData.VirtualTokenReserves) * 1000000
	solTokenPrice = math.Round(solTokenPrice*1e9) / 1e9

	TokenAmout := ((c.buyAmount) / (solTokenPrice))
	TokenAmoutInInt := uint64(TokenAmout * 1000000)
	solInWithSlippage := (c.buyAmount) * (1 + c.slippage)
	lamportsInWithSlippage := uint64(solInWithSlippage * 1000000000)

	createATAInstruction := associatedtokenaccount.NewCreateInstruction(
		Owner.PublicKey(),
		Owner.PublicKey(),
		c.mintAddress,
	).Build()

	accountKeys := solana.AccountMetaSlice{
		{PublicKey: solana.MustPublicKeyFromBase58("4wTV1YmiEkRvAtNtsSGPtUrqRYQMe5SKy2uB4Jjaxnjf"), IsSigner: false, IsWritable: false},
		{PublicKey: solana.MustPublicKeyFromBase58("CebN5WGQ4jvEPvsVU4EoHEpgzq1VV7AbicfhtW4xC9iM"), IsSigner: false, IsWritable: true},
		{PublicKey: c.mintAddress, IsSigner: false, IsWritable: false},
		{PublicKey: c.bondingCurveAddress, IsSigner: false, IsWritable: true},
		{PublicKey: c.associatedBondingCurve, IsSigner: false, IsWritable: true},
		{PublicKey: TokenAddress, IsSigner: false, IsWritable: true},
		{PublicKey: Owner.PublicKey(), IsSigner: true, IsWritable: true},
		{PublicKey: solana.SystemProgramID, IsSigner: false, IsWritable: false},
		{PublicKey: solana.TokenProgramID, IsSigner: false, IsWritable: false},
		{PublicKey: solana.SysVarRentPubkey, IsSigner: false, IsWritable: false},
		{PublicKey: solana.MustPublicKeyFromBase58("Ce6TQqeHC9p8KetsN6JsjHK7UTZk7nasjjnr7XxXp9F1"), IsSigner: false, IsWritable: false},
		{PublicKey: solana.MustPublicKeyFromBase58("6EF8rrecthR5Dkzon8Nwu78hRvfCKubJ14M5uBEwF6P"), IsSigner: false, IsWritable: false},
	}

	var buf bytes.Buffer
	binary.Write(&buf, binary.LittleEndian, uint64(16927863322537952870))
	binary.Write(&buf, binary.LittleEndian, uint64(TokenAmoutInInt))
	binary.Write(&buf, binary.LittleEndian, lamportsInWithSlippage)
	data := buf.Bytes()

	BuyInstruction := solana.NewInstruction(
		solana.MustPublicKeyFromBase58("6EF8rrecthR5Dkzon8Nwu78hRvfCKubJ14M5uBEwF6P"),
		accountKeys,
		data,
	)

	computeBudgetInstruction := computeBudget.NewSetComputeUnitPriceInstruction(uint64(517000)).Build()
	computeBudgetInstruction2 := computeBudget.NewSetComputeUnitLimitInstruction(uint32(72000)).Build()

	// Create a transaction
	instructions := []solana.Instruction{
		computeBudgetInstruction,
		computeBudgetInstruction2,
		createATAInstruction,
		BuyInstruction,
	}

	blockHash, err := client.GetRecentBlockhash(context.Background(), rpc.CommitmentFinalized)
	if err != nil {
		return solana.Signature{}, fmt.Errorf("error getting recent blockhash: %v", err)
	}

	tx, err := solana.NewTransaction(instructions, blockHash.Value.Blockhash, solana.TransactionPayer(Owner.PublicKey()))
	if err != nil {
		return solana.Signature{}, fmt.Errorf("error creating transaction: %v", err)
	}

	privateKeyGetter := func(pubKey solana.PublicKey) *solana.PrivateKey {
		if pubKey == Owner.PublicKey() {
			return &Owner
		}
		return nil
	}

	tx.Sign(privateKeyGetter)

	opts := rpc.TransactionOpts{
		SkipPreflight:       false,
		PreflightCommitment: rpc.CommitmentFinalized,
	}

	txID, err := client.SendTransactionWithOpts(context.Background(), tx, opts)

	if err != nil {
		return solana.Signature{}, fmt.Errorf("error sending transaction: %v", err)
	}

	c.coinBought = TokenAmout
	return txID, nil
}

func (c *Coin) Sell(client rpc.Client, Owner solana.PrivateKey) (solana.Signature, error) {

	BondingCurveData, err := getBondingCurveInfos(&client, c.bondingCurveAddress)
	if err != nil {
		return solana.Signature{}, err
	}

	solTokenPrice := float64(float64(BondingCurveData.VirtualSolReserves/solana.LAMPORTS_PER_SOL)) / float64(BondingCurveData.VirtualTokenReserves) * 1000000
	solTokenPrice = math.Round(solTokenPrice*1e9) / 1e9

	TokenAmoutInInt := uint64(c.coinBought) * 1000000
	solOut := (c.coinBought * solTokenPrice)
	solOutWithSlippage := solOut * (1 - c.slippage)
	lamportsOutWithSlippage := uint64(solOutWithSlippage * 1000000000)

	accountKeys := solana.AccountMetaSlice{
		{PublicKey: solana.MustPublicKeyFromBase58("4wTV1YmiEkRvAtNtsSGPtUrqRYQMe5SKy2uB4Jjaxnjf"), IsSigner: false, IsWritable: false},
		{PublicKey: solana.MustPublicKeyFromBase58("CebN5WGQ4jvEPvsVU4EoHEpgzq1VV7AbicfhtW4xC9iM"), IsSigner: false, IsWritable: true},
		{PublicKey: c.mintAddress, IsSigner: false, IsWritable: false},
		{PublicKey: c.bondingCurveAddress, IsSigner: false, IsWritable: true},
		{PublicKey: c.associatedBondingCurve, IsSigner: false, IsWritable: true},
		{PublicKey: c.tokenAddress, IsSigner: false, IsWritable: true},
		{PublicKey: Owner.PublicKey(), IsSigner: true, IsWritable: true},
		{PublicKey: solana.SystemProgramID, IsSigner: false, IsWritable: false},
		{PublicKey: solana.MustPublicKeyFromBase58("ATokenGPvbdGVxr1b2hvZbsiqW5xWH25efTNsLJA8knL"), IsSigner: false, IsWritable: false},
		{PublicKey: solana.TokenProgramID, IsSigner: false, IsWritable: false},
		{PublicKey: solana.MustPublicKeyFromBase58("Ce6TQqeHC9p8KetsN6JsjHK7UTZk7nasjjnr7XxXp9F1"), IsSigner: false, IsWritable: false},
		{PublicKey: solana.MustPublicKeyFromBase58("6EF8rrecthR5Dkzon8Nwu78hRvfCKubJ14M5uBEwF6P"), IsSigner: false, IsWritable: false},
	}

	var buf bytes.Buffer
	binary.Write(&buf, binary.LittleEndian, uint64(12502976635542562355))
	binary.Write(&buf, binary.LittleEndian, uint64(TokenAmoutInInt))
	binary.Write(&buf, binary.LittleEndian, lamportsOutWithSlippage)
	data := buf.Bytes()

	SellInstruction := solana.NewInstruction(
		solana.MustPublicKeyFromBase58("6EF8rrecthR5Dkzon8Nwu78hRvfCKubJ14M5uBEwF6P"),
		accountKeys,
		data,
	)

	instructions := []solana.Instruction{
		SellInstruction,
	}

	blockHash, err := client.GetRecentBlockhash(context.Background(), rpc.CommitmentFinalized)
	if err != nil {
		return solana.Signature{}, err
	}

	tx, err := solana.NewTransaction(instructions, blockHash.Value.Blockhash, solana.TransactionPayer(Owner.PublicKey()))
	if err != nil {
		return solana.Signature{}, err
	}

	privateKeyGetter := func(pubKey solana.PublicKey) *solana.PrivateKey {
		if pubKey == Owner.PublicKey() {
			return &Owner
		}
		return nil
	}

	tx.Sign(privateKeyGetter)

	opts := rpc.TransactionOpts{
		SkipPreflight:       false,
		PreflightCommitment: rpc.CommitmentFinalized,
	}

	txID, err := client.SendTransactionWithOpts(context.Background(), tx, opts)

	if err != nil {
		return solana.Signature{}, err
	}

	return txID, nil
}
