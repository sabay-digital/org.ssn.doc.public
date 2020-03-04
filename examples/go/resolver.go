package resolver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/julienschmidt/httprouter"
	"github.com/sabay-digital/sdk.golang.ssn.digital/ssn"
	"github.com/sabay-digital/sdk.golang.ssn.digital/ssnclient"
)

type verifySignerRequest struct {
	Signer      string `json:"signer"`
	Ssn_account string `json:"ssn_account"`
}

func resolverHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	req := ssnclient.ResolverRequest{}
	resolve := ssnclient.ResolverResponse{
		Status: 200,
	}

	// Parse the request body
	if r.Header["Content-Type"][0] == "application/x-www-form-urlencoded" {
		// Extract the URL encoded values from the request body
		in, err := ioutil.ReadAll(r.Body)
		ssn.Log(err, "resolverHandler: Read request body")
		parse, err := url.ParseQuery(string(in))
		ssn.Log(err, "resolveHanlder: Parse request body")

		switch {
		// Check keys
		case len(parse["hash"]) == 0:
			resolve.Title = "Missing hash key/value pair from POST request"
			resolve.Status = 422
		case len(parse["signature"]) == 0:
			resolve.Title = "Missing signature key/value pair from POST request"
			resolve.Status = 422
		case len(parse["signer"]) == 0:
			resolve.Title = "Missing signer key/value pair from POST request"
			resolve.Status = 422
		case len(parse["ssn_account"]) == 0:
			resolve.Title = "Missing ssn_account key/value pair from POST request"
			resolve.Status = 422
		// Check values
		case len(parse["hash"][0]) == 0:
			resolve.Title = "Missing value for hash from POST request"
			resolve.Status = 422
		case len(parse["signature"][0]) == 0:
			resolve.Title = "Missing value for signature from POST request"
			resolve.Status = 422
		case len(parse["signer"][0]) == 0:
			resolve.Title = "Missing value for signer from POST request"
			resolve.Status = 422
		case len(parse["ssn_account"][0]) == 0:
			resolve.Title = "Missing value for ssn_account from POST request"
			resolve.Status = 422
		default:
			req.Hash = parse.Get("hash")
			req.Signature = parse.Get("signature")
			req.Signer = parse.Get("signer")
			req.Ssn_account = parse.Get("ssn_account")
		}
	} else {
		// Get the raw body and decode to JSON
		err := json.NewDecoder(r.Body).Decode(&req)
		switch {
		case ssn.Log(err, "resolverHandler: Decode raw body to JSON"):
			resolve.Title = "application/json decoding error"
			resolve.Status = 422
		case len(req.Hash) == 0:
			resolve.Title = "Missing hash"
			resolve.Status = 422
		case len(req.Signature) == 0:
			resolve.Title = "Missing signature"
			resolve.Status = 422
		case len(req.Signer) == 0:
			resolve.Title = "Missing signer"
			resolve.Status = 422
		case len(req.Ssn_account) == 0:
			resolve.Title = "Missing ssn_account"
			resolve.Status = 422
		}
	}

	if resolve.Status == 200 {
		// Step 1 - Verify Payment address: https://github.com/sabay-digital/org.ssn.doc.public/blob/master/tg/tg004.md#step-1---verify-the-payment-address-is-for-the-resolver
		pa := strings.Split(ps.ByName("pa"), "*")
		if len(pa) != 2 {
			resolve.Title = "Invalid payment address format"
			resolve.Status = 422
		} else {
			paDomain := strings.Split(pa[1], ".")
			if paDomain[0] != "demo-service" {
				resolve.Title = "Payment address not for this resolver"
				resolve.Status = 400
			} else {
				// Step 2 - Verify ssn_account is an existing TL and collect ccy's: https://github.com/sabay-digital/org.ssn.doc.public/blob/master/tg/tg004.md#step-2---verify-ssn_account-is-an-existing-trusted-payment-provider
				// Get PPs
				ppReq, _ := http.NewRequest(http.MethodGet, ssnAPI+"/accounts/"+serviceSK.Address(), nil)

				// Execute the request
				res, err := http.DefaultClient.Do(ppReq)
				ssn.Log(err, "resolverHandler: Get service account")
				defer res.Body.Close()

				// Read the request response
				body, _ := ioutil.ReadAll(res.Body)

				account := ssn.Account{}
				// Take the JSON apart
				json.Unmarshal(body, &account)

				ccy := []ssnclient.ResolverPaymentDetails{}
				for i := range account.Balances {
					if account.Balances[i].Is_authorized && account.Balances[i].Asset_issuer == req.Ssn_account {
						ccy = append(ccy, ssnclient.ResolverPaymentDetails{
							Asset_code: account.Balances[i].Asset_code,
						})
					}
				}
				// Not finding any asset codes indicates no trustline
				fmt.Println(ccy)
				if len(ccy) == 0 {
					resolve.Title = "Requestor is not authorized to view payment details"
					resolve.Status = 403
				} else {
					// Step 3 - Verify signature: https://github.com/sabay-digital/org.ssn.doc.public/blob/master/tg/tg004.md#step-3---verify-signature-is-valid
					sigVerified, err := ssnclient.VerifySignature(req.Hash, req.Signature, req.Signer, ssnAPI)
					if ssn.Log(err, "resolverHandler: Verify signature") {
						resolve.Title = "Requestor is not authorized to view payment details"
						resolve.Status = 403
					} else if !sigVerified {
						resolve.Title = "Requestor is not authorized to view payment details"
						resolve.Status = 403
					} else {
						// Step 4 - Verify signer is on SSN account: https://github.com/sabay-digital/org.ssn.doc.public/blob/master/tg/tg004.md#step-4---verify-signature-is-a-signer-on-ssn_account
						sigPresent, err := ssnclient.VerifySigner(req.Signer, req.Ssn_account, ssnAPI)
						if ssn.Log(err, "resolverHandler: Verify Signer") {
							resolve.Title = "Requestor is not authorized to view payment details"
							resolve.Status = 403
						} else if {
							resolve.Title = "Requestor is not authorized to view payment details"
							resolve.Status = 403
						} else {
							// At this point we can now prepare a valid response: https://github.com/sabay-digital/org.ssn.doc.public/blob/master/tg/tg004.md#build-the-response
							// If our payment address is related to a specific payment this is where we would pull the details from the DB
							resolve.Network_address = serviceSK.Address()
							resolve.Payment_type = "merchant"
							resolve.Service_name = "Sabay Demo Merchant"
							resolve.Details = &ssnclient.ResolverResponseDetails{
								Memo:         pa[0],
								Payment_info: "One time payment",
								Payment:      ccy, // This resolver only returns currency codes accepted not amounts
							}
						}
					}
				}
			}
		}
	}

	resp, err := json.Marshal(resolve)
	ssn.Log(err, "resolverHandler: Marshall response")

	// Write the API out
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resolve.Status)
	w.Write(resp)
}
