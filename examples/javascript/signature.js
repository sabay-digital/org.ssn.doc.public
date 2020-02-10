const StellarSdk = require('stellar-sdk')
const crypto = require('crypto')

function hash() {
    let mesg = 'Some message'

    // Hash
    const hash = crypto.createHash('sha256')
    hash.update(mesg)

    console.log(hash.toString('hex'))
}

function sign(hash) {
    // GDMG5Z4XL3CNGHK2GJD5TFIDRWRCBFVFV3WAFWFSBONWB6AKDODILHFZ
    let kp = StellarSdk.Keypair.fromSecret('SDWMABEXMMUVENWEB73FB3EQJB5QSKOYIBXDOXAE6A3NIHIYRUQJSWXY')

    // Sign
    sig = kp.sign(hash.digest('hex'))

    console.log(signature.toString('hex'))
}

function verify(hash, signature) {
    let kp =  StellarSDK.Keypair.fromPublicKey('GDMG5Z4XL3CNGHK2GJD5TFIDRWRCBFVFV3WAFWFSBONWB6AKDODILHFZ')
    
    // Verify signature
    if (kp.verify(Buffer.from(hash, 'hex'), Buffer.from(signature, 'hex')) == true) {
        console.log('Signature valid')
    } else {
        console.log('Signature invalid')
    }
}