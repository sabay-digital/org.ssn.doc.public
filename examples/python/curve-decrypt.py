#
# decrypt text using stellar EdDAS keys
#

import sys, getopt, binascii, json, libnacl, libnacl.public
from stellar_base.keypair import Keypair

# get the message
line = sys.stdin.readline()
data = json.loads(line)

sk = b'SCMEVUFPDAH7YOY6QFPTRQT3AXUXBHBYCLI7CEWVEF7IBJK6ZXKP6VV5'
pk = data["public_key"]

# make keypairs
kp1 = Keypair.from_seed(sk)
kp2 = Keypair.from_address(pk)

# convert to curve keys
curve_kp1_sk = libnacl.crypto_sign_ed25519_sk_to_curve25519(kp1.signing_key.to_bytes() )
curve_kp2_pk = libnacl.crypto_sign_ed25519_pk_to_curve25519(kp2.verifying_key.to_bytes() )

# setup decryption box
crypt_box = libnacl.public.Box(curve_kp1_sk, curve_kp2_pk)

# decode
data["details"] = json.loads ( crypt_box.decrypt( binascii.unhexlify( data["encrypted"] ) ).decode("utf-8") )
del data["encrypted"] 
print( json.dumps(data, indent=4, sort_keys=False) )
