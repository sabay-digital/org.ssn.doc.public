package demopp

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/julienschmidt/httprouter"
	"github.com/sabay-digital/sdk.golang.ssn.digital/ssn"
	"github.com/sabay-digital/sdk.golang.ssn.digital/ssnclient"
)

type preAuthResponse struct {
	Status         int    `json:"status"`
	Hash           string `json:"hash,omitempty"`
	PaymentAddress string `json:"payment_address,omitempty"`
	Title          string `json:"title,omitempty"`
}

func preauthChargeHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Extract the URL encoded values from the request body
	in, err := ioutil.ReadAll(r.Body)
	ssn.Log(err, "preauthChargeHandler: Read request body")
	req, err := url.ParseQuery(string(in))
	ssn.Log(err, "preauthChargeHandler: Parse URL encoded values")

	// Step 1 - Resolve Payment Address: https://github.com/sabay-digital/org.ssn.doc.public/blob/master/tg/tg001.md#step-1---resolve-the-payment-address-1
	paURL := paResolver + "/resolve/" + ps.ByName("pa")

	// Hash the URL
	paMesg := sha256.New()
	paMesg.Write([]byte(paURL))

	// Sign the hash
	paSig, err := kp.Sign(paMesg.Sum(nil))
	ssn.Log(err, "checkoutHandler: Sign message")

	// Hex encode for resolver
	paHash := hex.EncodeToString(paMesg.Sum(nil))
	paSignature := hex.EncodeToString(paSig)

	payment, err := ssnclient.ResolvePA(ps.ByName("pa"), paHash, paSignature, kp.Address(), assetIssuer, paResolver)
	if ssn.Log(err, "preauthChargeHandler: Resolve payment address") {
		// Error
		fmt.Println("ResolvePA function returned error")
	} else if len(payment.Details.Payment) != 1 {
		// Error - the payment address must resolve a single payment amount and currency to process a preauth charge
		fmt.Println("Invalid PA for pre-authorization")
	} else if len(payment.Details.Payment[0].Amount) == 0 || len(payment.Details.Payment[0].Asset_code) == 0 {
		// Error - the payment address must resolve a single payment amount and currency to process a preauth charge
		fmt.Println("Invalid PA for pre-authorization")
	} else {
		preauthAccount, err := ssnclient.IsPreauthorizedAccount(payment.Network_address, ssnAPI)
		if ssn.Log(err, "preauthChargeHandler: Is preauthorized account") {
			// Error
			fmt.Println("IsPreauthorizedAccount function returned error")
		}

		// Step 2 - Verify Request Signature: https://github.com/sabay-digital/org.ssn.doc.public/blob/master/tg/tg001.md#step-2---verify-the-requst-signature-is-valid-1
		reqURL := "https://" + r.Host + r.RequestURI

		// Hash the URL
		reqMesg := sha256.New()
		reqMesg.Write([]byte(reqURL))
		reqHash := hex.EncodeToString(reqMesg.Sum(nil))

		if reqHash == req.Get("hash") {
			sigVerified, err := ssnclient.VerifySignature(req.Get("hash"), req.Get("signature"), req.Get("public_key"), ssnAPI)
			if ssn.Log(err, "preauthChargeHandler: Verify signature") {
				// Error
				fmt.Println("VerifySignature function returned error")
			} else if !sigVerified {
				// Error
				fmt.Println("Signature invalid")
			} else {
				// Step 3 - Verify Trust: https://github.com/sabay-digital/org.ssn.doc.public/blob/master/tg/tg001.md#step-3---verify-the-trustline
				trusted, err := ssnclient.VerifyTrust(payment.Network_address, payment.Details.Payment[0].Asset_code, assetIssuer, ssnAPI)
				if ssn.Log(err, "preauthChargeHandler: Verify trust") {
					// Error
					fmt.Println("VerifyTrust function returned error")
				} else if !trusted {
					// Error
					fmt.Println("Trust invalid")
				} else {
					// Step 4 - Check existing preauthorization: https://github.com/sabay-digital/org.ssn.doc.public/blob/master/tg/tg001.md#step-4---payment-provider-checks-the-preauthorization-exists-and-moves-the-amount-to-escrow
					var existingPreauth []PreAuthorization
					db.Where("user_pk = ? AND service_pk = ?", req.Get("public_key"), preauthAccount).Find(&existingPreauth)
					fmt.Println(existingPreauth)

					if len(existingPreauth) == 1 {
						// Step 5 - Build and sign SSN payment: https://github.com/sabay-digital/org.ssn.doc.public/blob/master/tg/tg001.md#step-5---build-and-sign-the-payment-for-ssn-1
						// Build Txn
						envelope, err := ssnclient.CreatePayment(assetIssuer, payment.Network_address, payment.Details.Payment[0].Amount, payment.Details.Payment[0].Asset_code, assetIssuer, payment.Details.Memo, ssnAPI)
						ssn.Log(err, "preauthChargeHandler: Create payment")

						// Sign Txn
						txn, err := ssnclient.SignTxn(envelope, localSigner, "ssn_testing_network")
						ssn.Log(err, "payHandler: Locally sign txn")

						// Submit Txn
						hash, err := ssnclient.SubmitTxn(txn, ssnAPI)
						ssn.Log(err, "preauthChargeHandler: Submit txn")
						fmt.Println(hash)

						resp := preAuthResponse{
							Status:         200,
							Hash:           hash,
							PaymentAddress: ps.ByName("pa"),
						}

						out, _ := json.Marshal(&resp)
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(resp.Status)
						w.Write(out)
					}
				}
			}
		}
	}
}
