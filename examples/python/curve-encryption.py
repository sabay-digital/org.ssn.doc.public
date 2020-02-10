# keys used

# merchant
# pk=SDM6IPWTKH4PU274YUJTKCNSQCAY6IHP6LTM6FCKJS535GEWKXSZTOM2
# sk=GCW6G3QMSC6KIEKIMP3K5OF7COISMZIXYST3DV26ICJJYGZ3ZV3BELXW

# cashier
# pk=GDYFDML3VYZ2XBG3DOTGFAR7PXSKME5UVBVL3JFBOF7LXXOGUUQOG5WL
# sk=SCMEVUFPDAH7YOY6QFPTRQT3AXUXBHBYCLI7CEWVEF7IBJK6ZXKP6VV5

import sys, binascii, json, libnacl, libnacl.public
from stellar_base.keypair import Keypair

# load the keys , sk1 we own, pk2 we get from request
sk1 = b'SDM6IPWTKH4PU274YUJTKCNSQCAY6IHP6LTM6FCKJS535GEWKXSZTOM2'
pk2 = b'GDYFDML3VYZ2XBG3DOTGFAR7PXSKME5UVBVL3JFBOF7LXXOGUUQOG5WL'

header = """
{
  "network_address": "GDTXTOPAMGXDSNHAOVLGPMSNHG3XIYTKZ5OUNVCP6J6XJI5AGMHSK3EP", 
  "public_key": "GCW6G3QMSC6KIEKIMP3K5OF7COISMZIXYST3DV26ICJJYGZ3ZV3BELXW", 
  "asset_code": "USD", 
  "payment_type": "merchant"
}
"""

details = """
{ 
  "memo": "sub:1112:1m",
  "service_name": "SOYO",
  "payment_info": "1 month subscription fee",
  "payment": {
      "amount": 1.99, 
      "asset_code": "USD"
  }
}
"""

# make keypairs
kp1 = Keypair.from_seed(sk1)
kp2 = Keypair.from_address(pk2)

# convert to curve keys
curve_kp1_sk = libnacl.crypto_sign_ed25519_sk_to_curve25519(kp1.signing_key.to_bytes() )
curve_kp2_pk = libnacl.crypto_sign_ed25519_pk_to_curve25519(kp2.verifying_key.to_bytes() )

# encrypt message towards pk2 key
crypt_box = libnacl.public.Box(curve_kp1_sk, curve_kp2_pk)
crypt_msg = crypt_box.encrypt( details.encode('utf-8') )

data = json.loads(header)
data["encrypted"] = binascii.hexlify ( crypt_msg ).decode("utf-8")

# return json object
print ( json.dumps(data) )
