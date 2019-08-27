
SSN Cashiers
============

This document describes the function and services provides by payment providers (cashiers) on SSN.

<!-- @import "[TOC]" {cmd="toc" depthFrom=1 depthTo=6 orderedList=false} -->

<!-- code_chunk_output -->

- [Nomenclature](#nomenclature)
- [Introduction to Cashiers](#introduction-to-cashiers)
- [Types of cashiers](#types-of-cashiers)
  - [One time payment](#one-time-payment)
  - [Pre-authorized payment](#pre-authorized-payment)
  - [Recurring payment](#recurring-payment)
  - [Direct Bill Payment](#direct-bill-payment)
- [Implementing a public cashier](#implementing-a-public-cashier)
- [Implementing authorization](#implementing-authorization)
  - [Pre-authorization](#pre-authorization)
  - [Recurring payment](#recurring-payment-1)
- [Verifying signatures](#verifying-signatures)
- [Payment process](#payment-process)
  - [Making transactions on SSN](#making-transactions-on-ssn)
  - [Payment address resolver](#payment-address-resolver)

<!-- /code_chunk_output -->

# Nomenclature 

**Cashiers** on SSN are asset issuing accounts, representing a payment provider which backs the assets issued via a FIAT reserve, and is licensed to hold money in escrow for user.

**Token** a so called "I owe you", refers to an asset issued on SSN by payment providers

**Asset** are tokens on SSN identified by a asset_code and asset_issuer pair.

**Accounts** are electronic wallets which can issue and hold tokens. Accounts are identified by a public key, see also *network address*

**Payment Address** a address which describes a payment destination (ex. invoice*service.example), the address must be resolved into a network address for a transaction on SSN , payment address encode service specific information for routing a payment to the correct account at the merchant or service provider.

**Network Address** are address describing an account on SSN using the public key of the account. *ex. GBIUFPAZAQDUI4U7E4476ARDWYCVAVUE7U7FRMRBQ72D6BRU6JYOVY43*

**Transactions** are operations on accounts, the operation can be between several accounts e.g. a transaction transferring tokens, or a transaction can be used to change settings on one account

**SSN Keys** are keys to sign transactions and to identify accounts (network address), each key comprises of a public key and a secret key. 

**Public Key** is used to identify accounts and authorized signers, a public key can be shared, several public keys (with different access authority) can be added to an account to sign transactions and/or change account settings

**Secret Key** is used to sign transactions which perform operations on SSN, secret keys should never be shared with anyone 

**Trustline** is a setting on a account to trust an asset issued by a cashier, only when an account trusts an asset, can the account hold this asset.

**Trust Line Authorization Required** requires that the issuing account approves a Trustline before the receiving account is allowed to be paid with the asset.

**Settlement Provider** is a service provider on SSN for settling balances between merchants and payment provider 

# Introduction to Cashiers

Cashiers on SSN 

* provide payment services to merchants 
* by authorizing a payment request with a user

and 

* recording the authorization on SSN 
* by issuing tokens on SSN which represent the payment. 

Tokens are a digital assets which are identified by the asset name (e.g. USD represents a USD backed asset) and the asset issuer (e.g. the institution which backs the asset with a underlying reserve). Each token issued is uniquely linked to a asset issuing account. Once the tokens are returned to that account, the tokens are destroyed. Cashiers sign the digital token creation on SSN via their own secret key, cryptographically verifying and securing the origin of the tokens.

Tokens must be backed by a underlying FIAT reserve held in escrow with the cashier. Cashiers are responsible to redeem tokens for the underlying FIAT asset at the agreed rate with a settlement provider when the tokens are returned to the cashier.

**Once a payment authorization is recorded, it is immutable.** It can not be changed, reversed or removed.

Payment provider can be a bank, a 3rd party payment processor, a mobile wallet provider or mobile carrier. All cashiers operating on SSN are listed in the asset directory under the cashier.ssn.digital domain.

Merchants on SSN, can configure their accounts to trust cashier and assets they issue, cashiers can also configure assets to require approval by the cashier before a merchant account can hold the asset.

During the payment process, the cashier must deduct the amount from the user balance and hold it in escrow until settlement with either the merchant directly or a settlement provider. When tokens are returned to the cashier account, the cashier must exchange them for the underlying fiat asset, minus fees and commissions. 

Settlement providers may help merchants in the settlement process by buying all tokens from a merchant and then settle with payment providers the consolidated amount.

***Cashiers shall publish the rates at with they will exchange the tokens issued for the underlying FIAT asset and may not settle below this rate.***

The process for a cashier involves the following steps upon receiving a payment request:

* verify the user
* resolve the payment address to get payment details
* use internal systems to deduct the users funds and moves them to an escrow account
* issue a transaction to SSN
* inform user of the transaction success with the SSN transaction ID
* inform merchant of the payment (if requested)

# Types of cashiers

A payment provider can implement different types of cashier services on SSN:

* one time payments (public endpoint)
* recurring payments (public endpoint)
* pre-authorized payments (public API)
* direct bill payments (no external public access)

A payment provide can implement all 4 solutions. It is advisable to have different asset issuing accounts for one-time and recurring charges due to different configuration which may apply to prevent abuse.

## One time payment

A one time payment is initiated by the cashier when the corresponding endpoint is called. The cashier must authenticate the user or provide a environment where the user is already authenticated (e.g. deep-link to a mobile application) and ask the user to agree to the charge.

## Pre-authorized payment

When the API is called the user is identified to the cashier via the pre-authorization key and ownership is verified with signatures. The cashier will execute the charge instantly, without asking the user for authorization. The service may inform the user about to charge by sending notification via a different channel.

To execute a pre-authorized payment, the merchant must have registered the user with the cashier and registered the authorization keys. 

## Recurring payment

For recurring payments the cashier needs to implement a solution whereby the same charge is execute based on the period settings in the request. During the setup the cashier must ask the user for authorization and confirm the recurring charge. On the charing date the cashier must issue the tokens on SSN and book the amount from the user balance to the escrow account (if the user has the required funds in his wallet).

Depending on the particulars of the payment provider, the charge may not occurs always on the same date. For example a recurring charge is setup with a mobile wallet provider, but the user did not top-up the wallet on time. In this case transaction maybe execute at a later date when the user tops up his wallet. 

The service provider is responsible to take delayed recurring payments into considerations when implementing the charge.

Asset issuing account for recurring payment should be configured with **Trust Line Authorization Required**.

## Direct Bill Payment

Direct Bill payment cashiers integrate the SSN blockchain API into their own platform, for example Online Banking Interface or Mobile Application without providing a public interface to access the cashier. 

Within the payment providers environment the user is prompted to provide the payment address **bill_id*merchant.example** (scanning QR codes, or past the payment address), and the payment provider will issue the transaction on SSN and move the balance to the settlement account.

The SSN PA resolver API will resolve the payment address and query the merchant for the billing amount and other details.

# Implementing a public cashier

The API reference to implemented a cashier is published at https://api-reference.ssn.digital. The minimum required implementation is the information service and the onetime charge API:

* `/status` a status end point which confirms the cashier service is ready to process transactions
* `/info` show the configuration of the cashier, assets issued, transaction limits
* `/charge/onetime/{payment_address}` a endpoint for payment authorization

The cashier may also implement support for pre-authorization by implementing the following endpoints:

* `/authorize/{public_key}/{ssn_account}` endpoint to setup pre-authorization for payments 
* `/charge/auth/{payment_address}` endpoint to execute pre-authorized charge

and recurring payments by implementing:

* `/recurring/{payment_address}` endpoint to setup recurring payments
* and the corresponding recurring charging mechanism's 

# Implementing authorization

The SSN's cashier reference implementation contains 2 authorization request where the cashier request authorization from the user to  

1) add pre-authorize payment keys
2) setup recurring charges from the users balance

## Pre-authorization

SSN is using public keys and signatures to implement pre-authorization of payments, meaning no user interaction is required to process the payment. Pre-authorization is implemented by public key signatures using EdDSA (Ed25519) keys. 

The merchant or service will assign the user a key which is comprised of a public key and private key, the service will store the private key with the user profile. The public key will be send to the cashier with the request to add the key to the user profile of the authenticated user. The user can then be identified by the cashier with this public key. The cashier may enforce transaction limits for pre-authorized payments and publish them in its info profile.

The API to implement the service is described in SSN 3rd party API for the endpoint

 `https://api.ssn.digital/v1​/authorize​/{public_key}​/{ssn_account}`

After the registration of the key, the merchant can construct a payment request on behalf of the user and use the users private key to sign the request. The cashier can verify the signature and the payment destination and execute the payment.

A user profile with a cashier may have more then one public key stored, if the user has pre-authorized payments with different merchants or service providers.

**Flow**

The cashier receives a request to add a public key to its user database to permit pre-authorized payments

1. authenticate the user and ask for permission for pre-authorization
2. store the public key with the user settings 
3. redirect to caller based header callback URL

## Recurring payment

For recurring payments, the cashier need to store the details of the request and periodically execute the same payment on behalf of the user. The cashier may implement a service to allow his user to manage the recurring payments via the cashiers own service interface, e.g. suspend the payments or cancel the payments. 

The service provider must accept a missing payment as cancellation of the service within an accepted grace period, if payments are restarted at a later period in time, the service provider needs to re-activate the service.

Recurring payment on SSN are represented in frequency, daily, weekly, monthly, yearly with a numeric qualifier for the duration.

  ```yaml
  recurring_payment_frequency: day
  recurring_payment_duration: 7
  recurring_payment_start: YYYY-MM-DD
  ```

if the recurring_payment_int is not present, a value of 1 should be assumed. If recurring_payment_start is not present, todays date should be used.

The cashier can implement the first charge with the setup of the recurring payment or the charging process maybe handled by a different service implemented at the cashier. Information about the recurring payment can either be submitted by the merchant in the request or resolved from the payment address.

**Flow**

1. resolve payment address from service provider (see SSN Payment Address API)
2. Check and valid input received (including data received from step 1)
3. verify if network account is able to receive the payment (validate trustline, see SSN 3rd party API)
4. authenticate the user and ask to authorize the recurring payment
5. record the user details in your recurring billing platform
    * authenticated user
    * payment address
    * recurring intervale 
    * next charge date
6. redirect to caller based header callback URL


# Verifying signatures

If the cashier receives a request with a authorization header, the cashier will need to verify if the signature is valid. For authorization using signatures a sha256 hash of complete URI will be signed.

Digital signatures can be checked in python with the following code:

```python {.line-numbers}
from stellar_base.keypair import Keypair
​
msg = b'https://cashier.provider.example/charge/auth/FB11212asq*m.ssn.digital'
keypair = Keypair.random()
​
print("Public key / Account address:\n",
      keypair.address().decode())
​
print("Seed / Secret key:\n",
      keypair.seed().decode())
​
sig = keypair.signing_key.sign(msg, encoding='hex')
print("Signature:\n",sig.decode())
​
try:
    keypair.verifying_key.verify(sig, msg, encoding='hex')
    print("The signature is valid.")
except:
    print("Invalid signature!")
```

The hash, signature and public key is transmitted in the header of the request using to following header-keys

* **X-SSN-hash** hash of the URI
* **X-SSN-signature** signature for the hash
* **X-SSN-public-key** public key which signed the hash

The cashier can verify that the secret key linked to the public key stored with the cashier has signed the request. The SSN 3rd party API has a function to valid signatures, however the cashier needs to validate that the hash corresponds to the URI using the sha256 hashing function.

The same request validation using signature schema, is also implement for revoking pre-authorization and revoking recurring charges.

# Payment process

The cashier receives the billing request either:

* via its public API, 
* from a recurring charge driver, 
* from a scanned QR code or 
* via direct user input.

*The billing process assumes the user is already authenticated and has given permission for the charge.*

**The cashier flow:**

1. resolve the payment address

In order to process the charge the cashier first needs to resolve the payment address. When querying the resolver it will return the payment request details (some details maybe passed on via the request body if the cashier process is invoked via the public API. Please see the section on **payment address resolvers** below for more details. 

`https://pa.ssn.digital/v1/`

```json
 # query
 {
  "asset_issuer": "GBIAIBT6NYGKA5T54BM73VA5LDSOJXBN56WEBNBFU77FTF6YMV2SP3CF",
  "public_key": "GDOGR3WT45537H6HNXUCAYKFA65W3JUT663PD7DZLMT5J7WCECV245NA",
  "payment_address": "FSADQWER*koc.sabay.com"
 }
 # response
 {
  "network_address": "GDTXTOPAMGXDSNHAOVLGPMSNHG3XIYTKZ5OUNVCP6J6XJI5AGMHSK3EP",
  "public_key": "GC3YDNXFW4SWTB6EPIS2473U5GH6BBIIE6GVP5MW6NVMPV5LMUACIQCM",
  "memo": "FSADQWER",
  "payment": {
      "amount": 0.3, 
      "asset_code": "USD"
    }
 }
```

At a minium the following information must be returned:

* the network address to which the payment will be send
* a memo which helps the merchant to identify the payment

**If the network address or memo is missing or the resolver requests ends in error, the cashier must exit the process**, inform the user and inform the service using the call-back URL if provided.

If resolving the payment address and the data provided in the request body do not contain the charge amount and currency, the cashier must: 

* check with the cashier which currencies are accepted if the cashier offers different currencies
* ask the user for the amount

If the amount to pay returned from the resolver differs from the amount in the request, the amount returned from the resolver takes precedence, unless the charge was already defined for a recurring payment. 

After the information gathering process the cashier should have all data to construct a payment.


2. Check and validate input received for the payment 

Verify if the merchant account is able to receive a payment from this cashier.

The cashier can access the SSN 3rd party API to verify trust lines from merchants to cashiers vit the API by sending a request with the following request body:

`https://api.ssn.digital/v1/verify/trust`

```json
 # query
 {
  "account": "GBR2ACKDSJXDGD62IWIJTPYUIKDMGEIRADZSIVZTT5G5MSDDM4732CQK",
  "asset_code": "USD",
  "asset_issuer": "GBIUFPAZAQDUI4U7E4476ARDWYCVAVUE7U7FRMRBQ72D6BRU6JYOVY43"
 }
 # response
 {
  "status": 200,
  "title": "asset will be accepted by account"
 }
```

Account is the merchant address one the network which will be checked for a trust-line to the asset_issuer (cashier) and asset_code combination. **If the trust line check fails the cashier must exit the payment process**, inform the user and inform the service using the call-back URL if provided.

3. lock requested balance in users wallet

Internal process, based on providers flows.

4. record payment on SSN and receive validation

See section "Making transactions on SSN"

5. transfer balance from user to escrow account

Internal process, based on providers flows.

For recurring payments, the next charging date must be moved forward.

6. show user success message with SSN transaction ID

The cashier has at this stage received the SSN transaction ID, it must present it to the user in a way in which the user can easily capture the ID via screenshot (QR code) or via a link which the user can save.

The transaction ID proves to the user that the payment was done and he can use this ID in any dispute with the merchant to verify he made the payment.

7. redirect user back to callback URL if requested

If at any step in the process there is an error, the cashier must (if possible) show the error to the user and redirect him back to the callback URL once the user acknowledges the error. 

## Making transactions on SSN

Making transactions on SSN is supported by the 3rd party API transaction builder, it consist of the following steps:

1. build payment transaction using SSN transaction builder 
    * network address of merchant (as returned from PA resolver)
    * asset_code
    * asset_issuer network address
    * amount
    * memo (as provided by merchant)
2. sign transaction with a signature key (e.g. a key which has authorization to sign transactions on behalf of the asset issuer account)
3. submit transaction to SSN 3rd party API
4. SSN returns transaction ID or rejects transaction

Once the transaction ID is received it should be shown to the user, so he has a permanent record for the payment. The Sabay github repository contains examples how to build transaction using the transaction builder.

The cashier can also build the transaction using the stellar SDK as published by the SDF. 

## Payment address resolver

A SSN payment address has the following formate

`payment_information*merchant.example`

The left side of the '\*' represent details needed by the merchant or service provider to identify what the payment is for. This could be a invoice ID, a user ID, a service ID or any other identifer as chosen by the merchant. The right side of the '\*' denotes the merchants in a form of a domain name.

On SSN, merchants operate lookup services to resolve destination address, the destination address can be resolved by accessing a resolver at the SSN API `https://pa.ssn.digital/` or the merchants resolver directly.

In order to protect information which can be gained from the lookup, the information returned by the resolver is always encrypted toward the public key provided in the request.

To establish trust that the query came from an authorized source, the request contains the following information:

`https://pa.ssn.digital/v1/`

```json
{
    "asset_issuer": "GBIAIBT6NYGKA5T54BM73VA5LDSOJXBN56WEBNBFU77FTF6YMV2SP3CF",
    "public_key": "GDOGR3WT45537H6HNXUCAYKFA65W3JUT663PD7DZLMT5J7WCECV245NA210",
    "payment_address": "FSADQWER*koc.sabay.com"
}
```

The asset_issuer is the public key of the cashier account which is requesting the information to process the payment. The public_key is the key used for the encryption. The key must be a signing key on the cashiers account to ensure it belongs to the cashier.

The encryption is based on ECC encryption using Curve25519 keys derived from the SSN EdDSA keys. The message is encrypted using x25519-xsalsa20-poly1305 and encoded in hex.


```json
{
  "network_address": "GDTXTOPAMGXDSNHAOVLGPMSNHG3XIYTKZ5OUNVCP6J6XJI5AGMHSK3EP",
  "public_key": "GC3YDNXFW4SWTB6EPIS2473U5GH6BBIIE6GVP5MW6NVMPV5LMUACIQCM",
  "encrypted": "57e54c4daa4fb58ad6f2a6a1c82da19b7391d6545451b34cf2590560c9\
              83384ada210a865fc9ca5abcc626b58f6d9b2e69c67a2de19cbe18fe5c\
              c522bacdc3e363839aca19cba18a0d710e34b8c97af7e589d45b3d99cb\
              dedd65ff524f8ed4c82dfabf01b3d2e48bf0db6b1d09f9c72f9df3b01b1\
              8fe89c0690a33b0ae546308060d4875d051eee024954f901e8bece49fa0\
              a6133896a8cda3d9b6c4cb00cb26d3cc59b4c7524d906ee1e69beb7d3200\
              543b4491148d5960bfcecf9b16923b48ec6f39d764ba7326d6281aae1da0\
              902fdf3332f1611a31af6acf"
}
```

For ECC encryption and decryption `libsodium` supports all function in use on SSN.

Decryption example in python using the SDK from stellar to manage the address and `libnacl` (python `libsodium` wrapper) to handle the decryption.

```python {.line-numbers}
 import sys
 import binascii
 import libnacl
 import libnacl.public
 from stellar_base.keypair import Keypair

 # define keys 
 # sk is the private key of the public key we send to the API
 # pk is a public key which we get back from the response
 sk = b'SCPLKQOZ4Z7CRRC6C7PYYLE4J2R7QYWWNK6VFZR7W4NJOBKORHOB7C7J'
 pk = b'GC3YDNXFW4SWTB6EPIS2473U5GH6BBIIE6GVP5MW6NVMPV5LMUACIQCM'

 # make key pairs
 kp1 = Keypair.from_seed(sk)
 kp2 = Keypair.from_address(pk)

 # convert EdDAS keys to curve keys
 curve_sk = libnacl.crypto_sign_ed25519_sk_to_curve25519(
    kp1.signing_key.to_bytes() )
 curve_pk = libnacl.crypto_sign_ed25519_pk_to_curve25519(
    kp2.verifying_key.to_bytes() )

 # setup decryption box
 crypt_box = libnacl.public.Box(curve_sk, curve_pk)

 # get the encrypted message in hex
 line = sys.stdin.readline()

 # decode
 message = crypt_box.decrypt( binascii.unhexlify( line ) )
 print(message)
```

after decryption

```json
{
  "network_address": "GDTXTOPAMGXDSNHAOVLGPMSNHG3XIYTKZ5OUNVCP6J6XJI5AGMHSK3EP",
  "public_key": "GC3YDNXFW4SWTB6EPIS2473U5GH6BBIIE6GVP5MW6NVMPV5LMUACIQCM",
  "details": {
    "memo": "sub:1112:1w",
    "payment": {
      "amount": 0.3,
      "asset_code": "USD"
    }
  }
}
```