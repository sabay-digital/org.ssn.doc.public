package onetimecharge

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/julienschmidt/httprouter"
	"github.com/sabay-digital/sdk.golang.ssn.digital/ssn"
	"github.com/sabay-digital/sdk.golang.ssn.digital/ssnclient"
)

func onetimeChargeHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Extract the URL encoded values from the request body
	in, err := ioutil.ReadAll(r.Body)
	ssn.Log(err, "onetimeChargeHandler: Read request body")
	req, err := url.ParseQuery(string(in))
	ssn.Log(err, "onetimeChargeHandler: Parse URL encoded values")

	// Step 1 - Verify Request Signature: https://github.com/sabay-digital/org.ssn.doc.public/blob/master/tg/tg001.md#step-1---verify-the-requst-signature-is-valid
	// Build the request URL
	reqURL := "https://" + r.Host + r.RequestURI
	fmt.Println(reqURL)
	// Hash the request URL
	reqMesg := sha256.New()
	reqMesg.Write([]byte(reqURL))
	reqHash := hex.EncodeToString(reqMesg.Sum(nil))
	fmt.Println(reqHash)
	fmt.Println(req.Get("hash"))
	// The hash we have built should match the request hash
	if reqHash == req.Get("hash") {
		sigVerified, err := ssnclient.VerifySignature(req.Get("hash"), req.Get("signature"), req.Get("public_key"), ssnAPI)
		if ssn.Log(err, "onetimeChargeHandler: Verify signature") {
			// Error
			fmt.Println("VerifySignature function returned error")
		} else if !sigVerified {
			// Error
			fmt.Println("Signature invalid")
		} else {
			// Step 2 - Resolve Payment Address: https://github.com/sabay-digital/org.ssn.doc.public/blob/master/tg/tg001.md#step-2---resolve-the-payment-address
			paURL := paResolver + "/resolve/" + ps.ByName("pa")
			fmt.Println(paURL)

			// Build the request body for the resolver
			// Hash the URI
			paMesg := sha256.New()
			paMesg.Write([]byte(paURL))
			// Sign the hash
			paSig, err := kp.Sign(paMesg.Sum(nil))
			ssn.Log(err, "checkoutHandler: Sign message")
			// Hex encode for resolver
			paHash := hex.EncodeToString(paMesg.Sum(nil))
			paSignature := hex.EncodeToString(paSig)

			// Resolve
			payment, err := ssnclient.ResolvePA(ps.ByName("pa"), paHash, paSignature, kp.Address(), assetIssuer, paResolver)
			if ssn.Log(err, "onetimeChargeHandler: Resolve payment address") {
				// Error
				fmt.Println("ResolvePA function returned error")
			} else {
				// Step 3 - Verify Trust: https://github.com/sabay-digital/org.ssn.doc.public/blob/master/tg/tg001.md#step-3---optionally-check-the-trustline-is-valid
				trusted, err := ssnclient.VerifyTrust(payment.Network_address, payment.Details.Payment[0].Asset_code, assetIssuer, ssnAPI)
				if ssn.Log(err, "onetimeChargeHandler: Verify trust") {
					// Error
					fmt.Println("VerifyTrust function returned error")
				} else if !trusted {
					// Error
					fmt.Println("Trust invalid")
				} else {
					// Step 4 - Payment Provider authorization: https://github.com/sabay-digital/org.ssn.doc.public/blob/master/tg/tg001.md#step-4---user-authorizes-the-payment-payment-provider-moves-the-amount-to-escrow

					// At this point the payment provider's UI/UX should takeover to authorize the user and deduct the funds
					// UI/UX flow should take account of the resolved payment details as documented here: https://github.com/sabay-digital/org.ssn.doc.public/blob/master/tg/tg001.md#step-2---resolve-the-payment-address

					// Step 5 - Build and Sign SSN payment: https://github.com/sabay-digital/org.ssn.doc.public/blob/master/tg/tg001.md#step-5---build-and-sign-the-payment-for-ssn
					// Build Txn
					envelope, err := ssnclient.CreatePayment(assetIssuer, payment.Network_address, payment.Details.Payment[0].Amount, payment.Details.Payment[0].Asset_code, assetIssuer, payment.Details.Memo, ssnAPI)
					ssn.Log(err, "payHandler: Create payment")

					// Sign Txn
					txn, err := ssnclient.SignTxn(envelope, localSigner, "ssn_testing_network")
					ssn.Log(err, "payHandler: Locally sign txn")

					// Submit Txn
					hash, err := ssnclient.SubmitTxn(txn, ssnAPI)
					ssn.Log(err, "payHandler: Submit txn")

					// Payment complete
					fmt.Println(hash)
				}
			}
		}
	}
}
