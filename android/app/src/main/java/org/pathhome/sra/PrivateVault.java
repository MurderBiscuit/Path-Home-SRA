package org.pathhome.sra;

import java.nio.charset.StandardCharsets;
import java.security.GeneralSecurityException;
import java.security.SecureRandom;
import java.util.Arrays;
import javax.crypto.Cipher;
import javax.crypto.SecretKeyFactory;
import javax.crypto.spec.GCMParameterSpec;
import javax.crypto.spec.PBEKeySpec;
import javax.crypto.spec.SecretKeySpec;

/** Versioned authenticated envelope: PHV1 || salt || nonce || ciphertext/tag.
 * Keystore encryption is applied outside this envelope by MainActivity.
 */
final class PrivateVault {
    private static final byte[] MAGIC = "PHV1".getBytes(StandardCharsets.US_ASCII);
    private static final int SALT=16, HEADER=20, NONCE=12, TAG=16, MAX=16*1024*1024;
    private static final SecureRandom RNG=new SecureRandom();
    static boolean valid(char[] pass){int n=Character.codePointCount(pass,0,pass.length);return n>=15&&n<=256;}
    static byte[] newSalt(){byte[] s=new byte[SALT];RNG.nextBytes(s);return s;}
    static byte[] salt(byte[] raw)throws GeneralSecurityException{
        if(raw==null||raw.length<HEADER+NONCE+TAG||raw.length>MAX||
                !Arrays.equals(MAGIC,Arrays.copyOf(raw,MAGIC.length)))throw new GeneralSecurityException("Invalid vault envelope");
        return Arrays.copyOfRange(raw,4,HEADER);
    }
    static byte[] derive(char[] pass,byte[] salt)throws GeneralSecurityException{
        if(salt.length!=SALT||pass.length>512)throw new GeneralSecurityException("Invalid key parameters");
        PBEKeySpec spec=new PBEKeySpec(pass,salt,600000,256);
        try{return SecretKeyFactory.getInstance("PBKDF2WithHmacSHA256").generateSecret(spec).getEncoded();}
        finally{spec.clearPassword();}
    }
    static byte[] seal(byte[] key,byte[] salt,byte[] plain)throws GeneralSecurityException{
        if(key.length!=32||salt.length!=SALT||plain.length>MAX-HEADER-NONCE-TAG)throw new GeneralSecurityException("Invalid vault parameters");
        byte[] header=new byte[HEADER];System.arraycopy(MAGIC,0,header,0,4);System.arraycopy(salt,0,header,4,SALT);
        byte[] nonce=new byte[NONCE];RNG.nextBytes(nonce);
        Cipher c=Cipher.getInstance("AES/GCM/NoPadding");c.init(Cipher.ENCRYPT_MODE,new SecretKeySpec(key,"AES"),new GCMParameterSpec(128,nonce));c.updateAAD(header);
        byte[] sealed=c.doFinal(plain),out=new byte[HEADER+NONCE+sealed.length];
        System.arraycopy(header,0,out,0,HEADER);System.arraycopy(nonce,0,out,HEADER,NONCE);System.arraycopy(sealed,0,out,HEADER+NONCE,sealed.length);
        return out;
    }
    static byte[] open(byte[] key,byte[] raw)throws GeneralSecurityException{
        salt(raw);
        if(key==null||key.length!=32)throw new GeneralSecurityException("Invalid vault key");
        Cipher c=Cipher.getInstance("AES/GCM/NoPadding");c.init(Cipher.DECRYPT_MODE,new SecretKeySpec(key,"AES"),new GCMParameterSpec(128,raw,HEADER,NONCE));c.updateAAD(raw,0,HEADER);
        return c.doFinal(raw,HEADER+NONCE,raw.length-HEADER-NONCE);
    }
    static void wipe(byte[] b){if(b!=null)Arrays.fill(b,(byte)0);}
    private PrivateVault(){}
}
