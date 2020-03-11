package demopp

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

type AddPreauthResponse struct {
	Currencies       []string
	Service_user_key string
	Service_key      string
	Service_name     string
	Redirect         string
}

func addPreauthHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Extract the URL encoded values from the request body
	in, err := ioutil.ReadAll(r.Body)
	ssn.Log(err, "addPreauthHandler: Read request body")
	req, err := url.ParseQuery(string(in))
	ssn.Log(err, "addPreauthHandler: Parse URL encoded values")

	// Step 1 - Verify Request Signature: https://github.com/sabay-digital/org.ssn.doc.public/blob/master/tg/tg001.md#step-1---verify-the-requst-signature-is-valid
	// Build the request URL
	reqURL := "https://" + r.Host + r.RequestURI
	// Hash the request URL
	reqMesg := sha256.New()
	reqMesg.Write([]byte(reqURL))
	reqHash := hex.EncodeToString(reqMesg.Sum(nil))
	// The hash we have built should match the request hash
	if reqHash == req.Get("hash") && req.Get("public_key") == ps.ByName("pk") {
		sigVerified, err := ssnclient.VerifySignature(req.Get("hash"), req.Get("signature"), req.Get("public_key"), ssnAPI)
		if ssn.Log(err, "addPreauthHandler: Verify signature") {
			// Error
			fmt.Println("VerifySignature function returned error")
		} else if !sigVerified {
			// Error
			fmt.Println("Signature invalid")
		} else {
			// Step 2 - Verify all trustlines in place: https://github.com/sabay-digital/org.ssn.doc.public/blob/master/tg/tg001.md#step-2---verify-the-trustlines
			// Should look up trust lines to MK and compare with PP customers accounts
			ccy := make([]string, 0)
			trust, err := ssnclient.VerifyTrust(ps.ByName("mk"), "USD", assetIssuer, ssnAPI)
			ccy = append(ccy, "USD")
			if ssn.Log(err, "addPreauthHandler: Verify trust") {
				// Error
				fmt.Println("VerifyTrust function returned error")
			} else if !trust {
				// Error
				fmt.Println("Trust missing")
			} else {
				// Step 3 - User authorizes the pre-authorization: https://github.com/sabay-digital/org.ssn.doc.public/blob/master/tg/tg001.md#step-3---user-authorizes-the-pre-authorization

				// At this point the payment provider's UI/UX should takeover to authorize the user and offer the ability to set pre-authorization limits
				// Get the merchant name from the network - this can be used in the view
				service, err := ssnclient.GetServiceName(ps.ByName("mk"), ssnAPI)
				ssn.Log(err, "addPreauthHandler: Get service name")

				// Step 4 - Add User and Merchant keys to a database: https://github.com/sabay-digital/org.ssn.doc.public/blob/master/tg/tg001.md#step-4---add-user-and-merchant-keys-to-a-database
			}
		}
	}
}
