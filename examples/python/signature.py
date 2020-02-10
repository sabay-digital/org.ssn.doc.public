from stellar_base.keypair import Keypair
import hashlib

def hash():
      mesg = "Some message"

      # Hash
      h = hashlib.sha256()
      h.update(mesg.encode())

      print(h.hexdigest())

def sign(hash):
      # GDMG5Z4XL3CNGHK2GJD5TFIDRWRCBFVFV3WAFWFSBONWB6AKDODILHFZ
      kp = Keypair.from_secret("SDWMABEXMMUVENWEB73FB3EQJB5QSKOYIBXDOXAE6A3NIHIYRUQJSWXY")

      # Sign
      signature = kp.signing_key.sign(hash, encoding='hex')

      print(signature.decode())


def verify(hash, signature):
      kp = Keypair.from_public_key("GDMG5Z4XL3CNGHK2GJD5TFIDRWRCBFVFV3WAFWFSBONWB6AKDODILHFZ")
 
      # Verify signature
      try:
            kp.verifying_key.verify(sig, msg, encoding='hex')
            print("Signature valid")
      except:
            print("Signature invalid")