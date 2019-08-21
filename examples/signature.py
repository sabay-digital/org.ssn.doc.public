# verify signature
# based on https://cryptobook.nakov.com/digital-signatures/eddsa-sign-verify-examples

from stellar_base.keypair import Keypair

msg = b'Message for Ed25519 signing'
keypair = Keypair.random()

print("Public key / Account address:\n",
      keypair.address().decode())

print("Seed / Your secret to keep it on local:\n",
      keypair.seed())
print("Seed / Secret key:\n",
      keypair.seed().decode())

sig = keypair.signing_key.sign(msg, encoding='hex')
print("Signature:\n",sig.decode())

try:
    keypair.verifying_key.verify(sig, msg, encoding='hex')
    print("The signature is valid.")
except:
    print("Invalid signature!")
