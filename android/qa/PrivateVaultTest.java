package org.pathhome.sra;
import java.nio.charset.StandardCharsets;
import java.security.GeneralSecurityException;
import java.util.Arrays;
public final class PrivateVaultTest {
    private static void require(boolean yes,String detail){if(!yes)throw new AssertionError(detail);}
    private static void reject(byte[] key,byte[] raw)throws Exception{
        try{PrivateVault.open(key,raw);throw new AssertionError("Invalid vault was accepted");}
        catch(GeneralSecurityException expected){}
    }
    public static void main(String[] args)throws Exception{
        char[] pass="synthetic phrase for vault tests".toCharArray(),wrong="different synthetic passphrase".toCharArray();
        require(PrivateVault.valid(pass)&&!PrivateVault.valid("123456".toCharArray()),"passphrase policy");
        byte[] salt=PrivateVault.newSalt();require(!Arrays.equals(salt,PrivateVault.newSalt()),"salt uniqueness");
        byte[] key=PrivateVault.derive(pass,salt),bad=PrivateVault.derive(wrong,salt);
        Arrays.fill(pass,'\0');Arrays.fill(wrong,'\0');
        byte[] plain="{\"note\":\"synthetic private plan\"}".getBytes(StandardCharsets.UTF_8);
        try{
            byte[] a=PrivateVault.seal(key,salt,plain),b=PrivateVault.seal(key,salt,plain);
            require(!Arrays.equals(a,b),"fresh nonce");
            require(Arrays.equals(PrivateVault.open(key,a),plain),"round trip");
            require(!new String(a,StandardCharsets.ISO_8859_1).contains("synthetic private plan"),"plaintext not present");
            reject(bad,a);reject(null,a);
            for(int i:new int[]{0,4,20,a.length-1}){byte[] tamper=a.clone();tamper[i]^=1;reject(key,tamper);}
            for(int n:new int[]{0,4,20,47})reject(key,Arrays.copyOf(a,n));
            System.out.println("PASS: Android Private Vault cryptography, wrong keys, tampering, nonce uniqueness and malformed envelopes");
        }finally{PrivateVault.wipe(key);PrivateVault.wipe(bad);PrivateVault.wipe(plain);}
    }
}
