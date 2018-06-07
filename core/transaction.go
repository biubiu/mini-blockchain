package core

import (
	"bytes"
	"crypto/md5"
	"crypto/rsa"
	"errors"
	"fmt"

	"../config"
	"../util"
)

type TransactionInput struct {
	PrevtxMap   [config.HashSize]byte
	OutputIndex uint32
	Signature   []byte
}

type TransactionOutput struct {
	Value   uint64
	Address rsa.PublicKey
}

/*
 * A transaction contains a list of Inputs and Outputs.
 * To become a valid transation, it must contain the all signatures
 * from all users
 */
type Transaction struct {
	Inputs  []TransactionInput
	Outputs []TransactionOutput
}

func CreateTransaction(ninput int, noutput int) Transaction {
	var tran Transaction
	for i := 0; i < ninput; i++ {
		var input TransactionInput
		tran.Inputs = append(tran.Inputs, input)
	}
	for i := 0; i < noutput; i++ {
		var output TransactionOutput
		tran.Outputs = append(tran.Outputs, output)
	}
	return tran
}

/*
 * Get the raw data to sign.
 * Basically it seralizes the transaction except the signature field
 */
func (tran *Transaction) getRawDataToSign() []byte {
	var data []byte
	for i := 0; i < len(tran.Inputs); i++ {
		data = appendUint32(data, tran.Inputs[i].OutputIndex)
		data = append(data, tran.Inputs[i].PrevtxMap[:]...)
	}

	for i := 0; i < len(tran.Outputs); i++ {
		data = appendUint64(data, tran.Outputs[i].Value)
		data = appendAddress(data, &tran.Outputs[i].Address)
	}
	return data
}

/*
 * Get the raw data to hash the whole transaction
 */
func (tran *Transaction) GetRawDataToHash() []byte {
	var data []byte
	for i := 0; i < len(tran.Inputs); i++ {
		data = appendUint32(data, tran.Inputs[i].OutputIndex)
		data = append(data, tran.Inputs[i].PrevtxMap[:]...)
		data = append(data, tran.Inputs[i].Signature...)
	}

	for i := 0; i < len(tran.Outputs); i++ {
		data = appendUint64(data, tran.Outputs[i].Value)
		data = appendAddress(data, &tran.Outputs[i].Address)
	}
	return data
}

/*
 * Get the raw data to hash the whole transaction
 */
func (tran *Transaction) GetRawDataToHashForTest() []byte {
	return tran.GetRawDataToHash()
}

/*
 * Sign a transaction in place (in practice, it should be called by each signer individually)
 */
func (tran *Transaction) SignTransaction(signers []*rsa.PrivateKey) error {
	if len(signers) != len(tran.Inputs) {
		return errors.New("Number of signers mismatch that of Inputs")
	}
	data := tran.getRawDataToSign()
	for i := 0; i < len(signers); i++ {
		signature, err := util.Sign(data, signers[i])
		if err != nil {
			return err
		}
		tran.Inputs[i].Signature = signature
	}
	return nil
}

/*
 * Verify whether a transaction has valid signatures.
 * Note that it doesn't verify whether the transaction is valid in the chain.
 */
func (tran *Transaction) VerifyTransaction(inputAddresses []*rsa.PublicKey) error {
	if len(inputAddresses) != len(tran.Inputs) {
		return errors.New("Number of Addresses mismatch that of Inputs")
	}
	data := tran.getRawDataToSign()
	for i := 0; i < len(inputAddresses); i++ {
		err := util.VerifySignature(data, tran.Inputs[i].Signature, inputAddresses[i])
		if err != nil {
			return err
		}
	}
	return nil
}

func (input TransactionInput) Print() string {
	return fmt.Sprintf("TransactionInput:%s[PrevtxMap:%s,OutputIndex:%x,Signature:%x],",
		util.Hash(input),
		util.HashBytes(input.PrevtxMap),
		input.OutputIndex,
		md5.Sum(input.Signature),
	)
}

func (output TransactionOutput) Print() string {
	return fmt.Sprintf("TransactionOutput:%s[Address:%v,Value:%v],",
		util.Hash(output),
		util.GetShortIdentity(output.Address),
		output.Value,
	)
}

func (tran Transaction) Print() string {
	var buffer bytes.Buffer
	for _, in := range tran.Inputs {
		buffer.WriteString(in.Print())
	}

	for _, out := range tran.Outputs {
		buffer.WriteString(out.Print())
	}

	return fmt.Sprintf("Transaction:%s[%s],", util.Hash(tran), buffer.String())
}
