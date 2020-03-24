import org.stellar.sdk.*;
import com.google.common.hash.*;
import com.google.common.io.*;

public class Signature {

    public static void hash() {
        // Usually a full request URL like https://pa.ssn.digital/v2/resolve/37837941*testing.mysabay.com
        String mesg = "Some message";
 
        // Hash
        HashCode h = Hashing.sha256().hashString(mesg, StandardCharsets.UTF_8);

        // Encode results in hex
        String hash = h.toString();
        
        System.out.println(hash);
    }

    public static void sign(String hash) {
        // GDMG5Z4XL3CNGHK2GJD5TFIDRWRCBFVFV3WAFWFSBONWB6AKDODILHFZ
        KeyPair kp = KeyPair.fromSecretSeed("SDWMABEXMMUVENWEB73FB3EQJB5QSKOYIBXDOXAE6A3NIHIYRUQJSWXY");

        byte[] mesg = BaseEncoding.base16().decode(hash); // Use guava.dev

        // Sign
        byte[] sig = kp.sign(mesg);

        String signature = BaseEncoding.base16().encode(sig); // Use guava.dev
        System.out.println(signature);
    }

    public static void verify(String hash, String signature) {
        KeyPair kp = KeyPair.fromAccountID("GDMG5Z4XL3CNGHK2GJD5TFIDRWRCBFVFV3WAFWFSBONWB6AKDODILHFZ");

        // Decode hash and message
        byte[] mesg = BaseEncoding.base16().decode(hash); // Use guava.dev
        byte[] sig = BaseEncoding.base16().decode(signature); // Use guava.dev

        if (kp.verify(mesg, sig) == true) {
            System.out.println("Signature valid");
        } else {
            System.out.println("Signature invalid");
        }
    }
}