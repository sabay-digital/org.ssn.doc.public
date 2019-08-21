#
# decrypt text using stellar EdDAS keys
#

import sys, getopt, binascii

import libnacl
import libnacl.public

from stellar_base.keypair import Keypair

sk = b'SCPLKQOZ4Z7CRRC6C7PYYLE4J2R7QYWWNK6VFZR7W4NJOBKORHOB7C7J'
pk = b'GC3YDNXFW4SWTB6EPIS2473U5GH6BBIIE6GVP5MW6NVMPV5LMUACIQCM'

# make keypairs
kp1 = Keypair.from_seed(sk)
kp2 = Keypair.from_address(pk)

# convert to curve keys
curve_kp1_sk = libnacl.crypto_sign_ed25519_sk_to_curve25519(kp1.signing_key.to_bytes() )
curve_kp2_pk = libnacl.crypto_sign_ed25519_pk_to_curve25519(kp2.verifying_key.to_bytes() )

# setup decryption box
crypt_box = libnacl.public.Box(curve_kp1_sk, curve_kp2_pk)

# get the message
line = sys.stdin.readline()

# decode
message = crypt_box.decrypt( binascii.unhexlify( line ) )
print(message)


def main(argv):

   # load the keys , sk1 we own, pk2 we get from request

   try:
      opts, args = getopt.getopt(argv,"hs:p:",["sk=","pk="])
   except getopt.GetoptError:
      print 'test.py -i <inputfile> -o <outputfile> < encrypted message'
      sys.exit(2)
   for opt, arg in opts:
      if opt == '-h':
         print 'test.py -s <secret_key> -p <public_key>'
         sys.exit()
      elif opt in ("-s", "--sk"):
         sk = arg
      elif opt in ("-p", "--pk"):
         pk = arg

if __name__ == "__main__":
   main(sys.argv[1:])
   decrypt(sk,pk)