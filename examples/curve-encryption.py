#
# encrypt text using SSN EdDAS keys with conversion to Curve25519 keys
#

import sys
import binascii

import libnacl
import libnacl.public

from stellar_base.keypair import Keypair

# load the keys , sk1 we own, pk2 we get from request
sk1 = b'SAUYKNIUDERGD73FERU7XCTK44WRQ4EPUUXEJK55XBWV3LBESSIE2BAQ'
pk2 = b'GDOGR3WT45537H6HNXUCAYKFA65W3JUT663PD7DZLMT5J7WCECV245NA'
message = b'{ \
  "network_address": "GDTXTOPAMGXDSNHAOVLGPMSNHG3XIYTKZ5OUNVCP6J6XJI5AGMHSK3EP", \
  "memo": "sub:1112:1w", \
  "payment": { \
      "amount": 0.3,  \
      "asset_code": "USD" \
    } \
}'

# make keypairs
kp1 = Keypair.from_seed(sk1)
kp2 = Keypair.from_address(pk2)

# convert to curve keys
curve_kp1_sk = libnacl.crypto_sign_ed25519_sk_to_curve25519(kp1.signing_key.to_bytes() )
curve_kp2_pk = libnacl.crypto_sign_ed25519_pk_to_curve25519(kp2.verifying_key.to_bytes() )

# encrypt message towards pk2 key
crypt_box = libnacl.public.Box(curve_kp1_sk, curve_kp2_pk)
crypt_msg = crypt_box.encrypt(message)

# message
sys.stdout.buffer.write( binascii.hexlify (crypt_msg) )
print ("\n")
