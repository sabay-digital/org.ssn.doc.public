package signature

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/stellar/go/keypair"
)

func hash() {
	mesg := "Some message"

	// Hash
	hash := sha256.New()
	hash.Write([]byte(mesg))

	fmt.Println(hex.EncodeToString(hash.Sum(nil)))
}

func sign(hash string) {
	// GDMG5Z4XL3CNGHK2GJD5TFIDRWRCBFVFV3WAFWFSBONWB6AKDODILHFZ
	kp := keypair.MustParse("SDWMABEXMMUVENWEB73FB3EQJB5QSKOYIBXDOXAE6A3NIHIYRUQJSWXY")

	mesg, _ := hex.DecodeString(hash)

	// Sign
	sig, _ := kp.Sign(mesg)

	fmt.Println(hex.EncodeToString(sig))
}

func verify(hash, signature string) {
	kp := keypair.MustParse("GDMG5Z4XL3CNGHK2GJD5TFIDRWRCBFVFV3WAFWFSBONWB6AKDODILHFZ")

	// Decode hash and message
	mesg, _ := hex.DecodeString(hash)
	sig, _ := hex.DecodeString(signature)

	// Verify signature
	err = kp.Verify(mesg, sig)
	if err != nil {
		fmt.Println("Signature invalid")
	} else {
		fmt.Println("Signature valid")
	}
}
