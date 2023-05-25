package main

import (
	"fmt"
	"os"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/consensys/gnark/backend/plonk"
	bn254plonk "github.com/consensys/gnark/backend/plonk/bn254"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/scs"
	"github.com/consensys/gnark/test"
	"github.com/consensys/plonk-solidity/tmpl"
)

func checkError(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

// ------------------------------------------
// no commitment
type noCommitmentCircuit struct {
	X frontend.Variable
	Y frontend.Variable `gnark:",public"`
}

func (c *noCommitmentCircuit) Define(api frontend.API) error {
	a := api.Mul(c.X, c.X, c.X, c.X)
	api.AssertIsEqual(a, c.Y)
	return nil
}

func getVkProofnoCommitmentCircuit() (bn254plonk.Proof, bn254plonk.VerifyingKey, []fr.Element) {

	var circuit noCommitmentCircuit
	ccs, err := frontend.Compile(ecc.BN254.ScalarField(), scs.NewBuilder, &circuit)
	checkError(err)

	var witness noCommitmentCircuit
	witness.X = 2
	witness.Y = 16
	witnessFull, err := frontend.NewWitness(&witness, ecc.BN254.ScalarField())
	checkError(err)
	witnessPublic, err := witnessFull.Public()
	checkError(err)

	srs, err := test.NewKZGSRS(ccs)
	checkError(err)

	pk, vk, err := plonk.Setup(ccs, srs)
	checkError(err)

	proof, err := plonk.Prove(ccs, pk, witnessFull)
	checkError(err)

	err = plonk.Verify(proof, vk, witnessPublic)
	checkError(err)

	tvk := vk.(*bn254plonk.VerifyingKey)
	tproof := proof.(*bn254plonk.Proof)

	ipi := witnessPublic.Vector()
	pi := ipi.(fr.Vector)

	return *tproof, *tvk, pi
}

// ------------------------------------------
// single commitment
type singleCommitmentCircuit struct {
	Public [3]frontend.Variable `gnark:",public"`
	X      [3]frontend.Variable
}

func (c *singleCommitmentCircuit) Define(api frontend.API) error {

	committer, ok := api.(frontend.Committer)
	if !ok {
		return fmt.Errorf("type %T doesn't impl the Committer interface", api)
	}
	commitment, err := committer.Commit(c.X[:]...)
	if err != nil {
		return err
	}
	for i := 0; i < 3; i++ {
		api.AssertIsDifferent(commitment, c.X[i])
		for _, p := range c.Public {
			api.AssertIsDifferent(p, 0)
		}
	}
	return err
}

func getVkProofsingleCommitmentCircuit() (bn254plonk.Proof, bn254plonk.VerifyingKey, []fr.Element) {

	var circuit singleCommitmentCircuit
	ccs, err := frontend.Compile(ecc.BN254.ScalarField(), scs.NewBuilder, &circuit)
	checkError(err)

	var witness singleCommitmentCircuit
	witness.X = [3]frontend.Variable{3, 4, 5}
	witness.Public = [3]frontend.Variable{6, 7, 8}
	witnessFull, err := frontend.NewWitness(&witness, ecc.BN254.ScalarField())
	checkError(err)
	witnessPublic, err := witnessFull.Public()
	checkError(err)

	srs, err := test.NewKZGSRS(ccs)
	checkError(err)

	pk, vk, err := plonk.Setup(ccs, srs)
	checkError(err)

	proof, err := plonk.Prove(ccs, pk, witnessFull)
	checkError(err)

	err = plonk.Verify(proof, vk, witnessPublic)
	checkError(err)

	tvk := vk.(*bn254plonk.VerifyingKey)
	tproof := proof.(*bn254plonk.Proof)

	ipi := witnessPublic.Vector()
	pi := ipi.(fr.Vector)

	return *tproof, *tvk, pi
}

// ------------------------------------------
// multiple commitments

type multipleCommitmentCircuit struct {
	X frontend.Variable
	Y frontend.Variable `gnark:",public"`
}

func (c *multipleCommitmentCircuit) Define(api frontend.API) error {

	a := api.Mul(c.X, c.X, c.X)

	committer, ok := api.(frontend.Committer)
	if !ok {
		return fmt.Errorf("type %T doesn't impl the Committer interface", api)
	}

	b, err := committer.Commit(a)
	if err != nil {
		return err
	}

	d, err := committer.Commit(b)
	if err != nil {
		return err
	}

	e, err := committer.Commit(a, b, d)
	if err != nil {
		return err
	}

	api.AssertIsDifferent(e, c.Y)

	return nil
}

func getVkProofmultipleCommitmentCircuit() (bn254plonk.Proof, bn254plonk.VerifyingKey, []fr.Element) {

	var circuit multipleCommitmentCircuit
	ccs, err := frontend.Compile(ecc.BN254.ScalarField(), scs.NewBuilder, &circuit)
	checkError(err)

	var witness multipleCommitmentCircuit
	witness.X = 2
	witness.Y = 3
	witnessFull, err := frontend.NewWitness(&witness, ecc.BN254.ScalarField())
	checkError(err)
	witnessPublic, err := witnessFull.Public()
	checkError(err)

	srs, err := test.NewKZGSRS(ccs)
	checkError(err)

	pk, vk, err := plonk.Setup(ccs, srs)
	checkError(err)

	proof, err := plonk.Prove(ccs, pk, witnessFull)
	checkError(err)

	err = plonk.Verify(proof, vk, witnessPublic)
	checkError(err)

	tvk := vk.(*bn254plonk.VerifyingKey)
	tproof := proof.(*bn254plonk.Proof)

	ipi := witnessPublic.Vector()
	pi := ipi.(fr.Vector)

	return *tproof, *tvk, pi
}

//go:generate go run main.go
func main() {

	// proof, vk, pi := getVkProofsingleCommitmentCircuit()
	proof, vk, pi := getVkProofsingleCommitmentCircuit()

	err := tmpl.GenerateVerifier(vk, proof, pi, "../contracts")
	checkError(err)

}
