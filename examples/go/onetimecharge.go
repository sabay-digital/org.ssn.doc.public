package onetimecharge

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"text/template"

	"github.com/julienschmidt/httprouter"
	"github.com/sabay-digital/sdk.golang.ssn.digital/ssn"
	"github.com/sabay-digital/sdk.golang.ssn.digital/ssnclient"
)

type OnetimeResponse struct {
	Payment_service_name string
	Payment_destination  string
	Payment_details      []ssnclient.ResolverPaymentDetails
	Payment_memo         string
	Redirect             string
}

func onetimeChargeHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Extract the URL encoded values from the request body
	in, err := ioutil.ReadAll(r.Body)
	ssn.Log(err, "onetimeChargeHandler: Read request body")
	req, err := url.ParseQuery(string(in))
	ssn.Log(err, "onetimeChargeHandler: Parse URL encoded values")

	// Verify Request Signature
	reqURL := "https://" + r.Host + r.RequestURI
	fmt.Println(reqURL)
	reqMesg := sha256.New()
	reqMesg.Write([]byte(reqURL))
	reqHash := hex.EncodeToString(reqMesg.Sum(nil))
	fmt.Println(reqHash)
	fmt.Println(req.Get("hash"))
	if reqHash == req.Get("hash") {
		sigVerified, err := ssnclient.VerifySignature(req.Get("hash"), req.Get("signature"), req.Get("public_key"), ssnAPI)
		if ssn.Log(err, "onetimeChargeHandler: Verify signature") {
			// Error
			fmt.Println("VerifySignature function returned error")
		} else if !sigVerified {
			// Error
			fmt.Println("Signature invalid")
		} else {
			// Resolve Payment Address
			paURL := paResolver + "/resolve/" + ps.ByName("pa")
			fmt.Println(paURL)

			// Hash the URI
			paMesg := sha256.New()
			paMesg.Write([]byte(paURL))

			// Sign the hash
			paSig, err := kp.Sign(paMesg.Sum(nil))
			ssn.Log(err, "checkoutHandler: Sign message")

			// Hex encode for resolver
			paHash := hex.EncodeToString(paMesg.Sum(nil))
			paSignature := hex.EncodeToString(paSig)

			payment, err := ssnclient.ResolvePA(ps.ByName("pa"), paHash, paSignature, kp.Address(), assetIssuer, paResolver)
			if ssn.Log(err, "onetimeChargeHandler: Resolve payment address") {
				// Error
				fmt.Println("ResolvePA function returned error")
			} else {
				redirect := req.Get("redirect")
				if len(req.Get("redirect")) == 0 {
					redirect = "/v1/success"
				}
				fmt.Println(redirect)

				resp := OnetimeResponse{
					Payment_service_name: payment.Service_name,
					Payment_destination:  payment.Network_address,
					Payment_details:      payment.Details.Payment,
					Payment_memo:         payment.Details.Memo,
					Redirect:             redirect,
				}

				paymentTemplate := template.Must(template.ParseFiles("templates/payment.html"))
				paymentTemplate.Execute(w, resp)
			}
		}
	}
}
