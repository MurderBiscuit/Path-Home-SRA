package org.pathhome.sra;

import android.app.Activity;
import android.app.DatePickerDialog;
import android.Manifest;
import android.content.pm.PackageManager;
import android.location.Location;
import android.location.LocationListener;
import android.location.LocationManager;
import android.os.Build;
import android.os.CancellationSignal;
import android.os.VibrationEffect;
import android.os.Vibrator;
import android.content.Intent;
import android.content.Context;
import android.graphics.Color;
import android.graphics.Typeface;
import android.graphics.drawable.GradientDrawable;
import android.os.Bundle;
import android.provider.CalendarContract;
import android.security.keystore.KeyGenParameterSpec;
import android.security.keystore.KeyProperties;
import android.view.Gravity;
import android.view.View;
import android.view.ViewGroup;
import android.widget.*;
import org.json.*;
import java.io.*;
import java.nio.charset.StandardCharsets;
import java.security.KeyStore;
import java.text.SimpleDateFormat;
import java.util.*;
import javax.crypto.Cipher;
import javax.crypto.KeyGenerator;
import javax.crypto.SecretKey;
import javax.crypto.spec.GCMParameterSpec;
import android.util.Base64;
import android.net.Uri;

/** Real native Android Views UI. Works offline; public legal reference data bundled locally.
 * Personal planning notes encrypted using Android Keystore and remain on this device.
 * Do not treat directory listings or research summaries as legal advice.
 */
public class MainActivity extends Activity {
    private static final int NAVY=Color.rgb(18,42,62), TEAL=Color.rgb(14,112,99), INK=Color.rgb(25,44,55), PAPER=Color.rgb(247,247,242), MUTED=Color.rgb(92,112,123), WHITE=Color.WHITE, CREAM=Color.rgb(255,246,225);
    private LinearLayout root, content, toolbar;
    private ScrollView scroll;
    private EditText countyInput;
    private JSONObject profile = new JSONObject(), legal = new JSONObject();
    private boolean profileLocked = false;
    private boolean vaultLocked=false, vaultBusy=false, foreground=false;
    private byte[] vaultKey, vaultEnvelope;
    private int vaultGeneration=0, vaultFailures=0;
    private long vaultNextAttempt=0;
    private String vaultNotice="";
    // Avoid encrypting and blocking disk writes on every character typed.
    private final android.os.Handler saveHandler = new android.os.Handler(android.os.Looper.getMainLooper());
    private boolean savePending = false;
    private final Runnable deferredSave = () -> {savePending=false;saveProfile();};
    private void queueSave(){if(profileLocked||vaultLocked)return;savePending=true;saveHandler.removeCallbacks(deferredSave);saveHandler.postDelayed(deferredSave,350L);}
    private void flushSave(){if(savePending){saveHandler.removeCallbacks(deferredSave);savePending=false;saveProfile();}}
    @Override protected void onPause(){flushSave();super.onPause();}
    @Override protected void onStart(){super.onStart();foreground=true;}
    @Override protected void onStop(){
        foreground=false;vaultGeneration++;vaultBusy=false;
        if(vaultKey!=null&&!vaultLocked){
            lockVault(false);
            // End the protected activity on backgrounding to destroy stale dialogs and views.
            // Rotation creates a fresh locked activity; other returns require reopening Path Home.
            if(!isChangingConfigurations()&&!isFinishing())finish();
        }
        super.onStop();
    }
    @Override protected void onDestroy(){vaultGeneration++;PrivateVault.wipe(vaultKey);vaultKey=null;saveHandler.removeCallbacks(deferredSave);super.onDestroy();}
    private String page = "home";
    private final Map<String,Integer> images = new HashMap<>();
    private final ArrayList<String> jurisdictions = new ArrayList<>();
    private final ArrayList<String> stateCodes = new ArrayList<>();
    private static final int LOCATION_REQUEST = 710;
    private static final long REMINDER_COOLDOWN_MS = 30L*60L*1000L;
    private boolean awaitingSaveLocation = false;
    private boolean manualCheck = false;
    private boolean checkInProgress = false;
    private static final String REMINDER_PROMPT = "Are you in the right place, and the right time? Just checking on you!";
    private static final String REMINDER_SUPPORT = "You've worked hard to get this far. Take a moment to think about where you're going and what matters most to you.";
    private static final String[] START_TASKS = {"Confirm first reporting date", "Birth certificate", "Social Security replacement", "State ID or driver license", "Confirm a place to stay", "Arrange reliable transportation", "Contact employment services", "Apply for food or healthcare help"};
    private static final String[] FAMILY_TASKS = {"Release day and pickup", "A safe place to sleep", "Birth certificate plan", "Social Security card plan", "ID or license plan", "Working phone", "Know the first reporting appointment", "Transportation to appointments", "Food and healthcare", "Employment or training", "Save funds for documents", "Ask what support is wanted"};

    @Override public void onCreate(Bundle b) {super.onCreate(b);
        getWindow().setStatusBarColor(NAVY);getWindow().setNavigationBarColor(NAVY);
        images.put("hero",R.drawable.hero);images.put("release_plan",R.drawable.release_plan);images.put("appointments",R.drawable.appointments);images.put("local_help",R.drawable.local_help);images.put("money",R.drawable.money);images.put("family",R.drawable.family);images.put("progress",R.drawable.progress);
        loadData();loadProfile();show(b==null?"home":b.getString("currentPage","home"));
        if(profileLocked){new android.app.AlertDialog.Builder(this).setTitle("Your private plan needs attention")
            .setMessage("The encrypted plan could not be opened. Your existing data was not replaced. Close the app and preserve the encrypted file.")
            .setCancelable(false).setPositiveButton("Close app",(d,w)->finish()).show();}
        if(legal.optJSONArray("jurisdictions")==null)toast("Legal reference data is unavailable. Do not rely on results from this copy.");
    }
    @Override protected void onSaveInstanceState(Bundle out){out.putString("currentPage",page);super.onSaveInstanceState(out);}
    @Override public void onBackPressed(){if(vaultLocked||profileLocked){finish();return;}if(!"home".equals(page))show("home");else super.onBackPressed();}
    private void loadData(){stateCodes.add("");jurisdictions.add("Choose your state — do not guess");try(InputStream f=getAssets().open("reentry-public.json")){
        ByteArrayOutputStream o=new ByteArrayOutputStream();byte[] chunk=new byte[8192];int n;while((n=f.read(chunk))!=-1)o.write(chunk,0,n);
        legal=new JSONObject(o.toString(StandardCharsets.UTF_8.name()));JSONArray a=legal.optJSONArray("jurisdictions");
        if(a!=null)for(int i=0;i<a.length();i++){JSONObject obj=a.optJSONObject(i);if(obj!=null&&!"FED".equals(obj.optString("jurisdiction_code"))){stateCodes.add(obj.optString("jurisdiction_code"));jurisdictions.add(obj.optString("jurisdiction_name"));}}
    }catch(Exception ignored){legal=new JSONObject();}}
    private void loadProfile(){
        try{
            String enc=getPreferences(MODE_PRIVATE).getString("private_plan",null);
            if(enc==null)return;
            String payload=decrypt(enc);
            if(payload.startsWith("PHV1:")){
                byte[] raw=Base64.decode(payload.substring(5),Base64.NO_WRAP);
                PrivateVault.salt(raw);vaultEnvelope=raw;vaultLocked=true;
                getWindow().addFlags(android.view.WindowManager.LayoutParams.FLAG_SECURE);
            }else profile=new JSONObject(payload);
        }catch(Exception e){profileLocked=true;profile=new JSONObject();}
    }
    private boolean persistProfile(){
        if(profileLocked||vaultLocked){toast("Encrypted plan is locked. No information was overwritten.");return false;}
        byte[] plain=null;
        try{
            String payload;byte[] next=null;
            if(vaultEnvelope!=null){
                if(vaultKey==null)throw new IOException("Missing vault key");
                plain=profile.toString().getBytes(StandardCharsets.UTF_8);
                next=PrivateVault.seal(vaultKey,PrivateVault.salt(vaultEnvelope),plain);
                payload="PHV1:"+Base64.encodeToString(next,Base64.NO_WRAP);
            }else payload=profile.toString();
            String encrypted=encrypt(payload);
            if(!getPreferences(MODE_PRIVATE).edit().putString("private_plan",encrypted).commit())throw new IOException("Secure storage rejected the write");
            if(next!=null)vaultEnvelope=next;
            return true;
        }catch(Exception e){toast("Could not securely save data. Please check storage.");return false;}
        finally{PrivateVault.wipe(plain);}
    }
    private void saveProfile(){savePending=!persistProfile();}
    private EditText vaultPassword(LinearLayout dest,String label){
        add(dest,text(label,16,INK,true));
        EditText e=new EditText(this);e.setSingleLine(true);
        e.setInputType(android.text.InputType.TYPE_CLASS_TEXT|android.text.InputType.TYPE_TEXT_VARIATION_PASSWORD|android.text.InputType.TYPE_TEXT_FLAG_NO_SUGGESTIONS);
        e.setFilters(new android.text.InputFilter[]{new android.text.InputFilter.LengthFilter(256)});
        e.setImportantForAutofill(View.IMPORTANT_FOR_AUTOFILL_NO_EXCLUDE_DESCENDANTS);
        e.setSaveEnabled(false);add(dest,e);return e;
    }
    private char[] takePassword(EditText e){
        char[] out=new char[e.length()];e.getText().getChars(0,e.length(),out,0);e.getText().clear();return out;
    }
    private void deriveVaultAsync(char[] pass,byte[] salt,boolean enrollment){
        if(vaultBusy||!foreground){Arrays.fill(pass,'\0');return;}
        vaultBusy=true;final int generation=++vaultGeneration;
        new Thread(()->{
            byte[] key=null;
            try{key=PrivateVault.derive(pass,salt);}catch(Exception ignored){}finally{Arrays.fill(pass,'\0');}
            final byte[] derived=key;
            runOnUiThread(()->{
                if(generation!=vaultGeneration||!foreground||isFinishing()||isDestroyed()){PrivateVault.wipe(derived);return;}
                vaultBusy=false;
                if(derived==null){toast("Could not derive vault key.");return;}
                if(enrollment){
                    if(vaultEnvelope!=null||vaultLocked||profileLocked){PrivateVault.wipe(derived);return;}
                    byte[] marker;
                    try{marker=PrivateVault.seal(derived,salt,new byte[0]);}
                    catch(Exception e){PrivateVault.wipe(derived);toast("Could not enable Private Vault. Existing plan preserved.");return;}
                    vaultEnvelope=marker;vaultKey=derived;
                    if(!persistProfile()){
                        // SharedPreferences may have updated its in-memory cache despite commit failure.
                        // Stop rather than downgrade a possibly protected record on another write.
                        PrivateVault.wipe(vaultKey);vaultKey=null;profileLocked=true;profile=new JSONObject();savePending=false;
                        toast("Vault enrollment could not be confirmed. Close and reopen Path Home. Do not erase the saved plan.");
                        finish();return;
                    }
                    savePending=false;getWindow().addFlags(android.view.WindowManager.LayoutParams.FLAG_SECURE);
                    show("privacy");toast("Private Vault enabled. Store your passphrase safely; it cannot be reset.");
                    return;
                }
                byte[] raw=null;JSONObject next;
                try{
                    raw=PrivateVault.open(derived,vaultEnvelope);
                    next=new JSONObject(new String(raw,StandardCharsets.UTF_8));
                }catch(Exception e){
                    PrivateVault.wipe(derived);
                    vaultFailures++;vaultNextAttempt=android.os.SystemClock.elapsedRealtime()+Math.min(vaultFailures,15)*1000L;
                    toast("Incorrect passphrase or damaged vault. Saved data unchanged.");return;
                }finally{PrivateVault.wipe(raw);}
                profile=next;vaultKey=derived;vaultLocked=false;vaultFailures=0;vaultNextAttempt=0;vaultNotice="";
                show("home");
            });
        },"path-home-vault-kdf").start();
    }
    private void enableVault(EditText first,EditText second){
        char[] pass=takePassword(first),confirm=takePassword(second);
        boolean ok=PrivateVault.valid(pass)&&Arrays.equals(pass,confirm);
        Arrays.fill(confirm,'\0');
        if(!ok){Arrays.fill(pass,'\0');toast("Use matching passphrases of at least 15 characters.");return;}
        deriveVaultAsync(pass,PrivateVault.newSalt(),true);
    }
    private void unlockVault(EditText field){
        char[] pass=takePassword(field);
        if(android.os.SystemClock.elapsedRealtime()<vaultNextAttempt){Arrays.fill(pass,'\0');toast("Wait before trying again.");return;}
        try{deriveVaultAsync(pass,PrivateVault.salt(vaultEnvelope),false);}
        catch(Exception e){Arrays.fill(pass,'\0');toast("Vault damaged. Saved file preserved.");}
    }
    private void lockVault(boolean render){
        if(vaultLocked||vaultKey==null)return;
        saveHandler.removeCallbacks(deferredSave);savePending=false;
        if(!persistProfile())vaultNotice="Some unsaved changes could not be saved. The last encrypted plan is preserved.";
        PrivateVault.wipe(vaultKey);vaultKey=null;profile=new JSONObject();vaultLocked=true;
        awaitingSaveLocation=false;manualCheck=false;checkInProgress=false;
        vaultGeneration++;vaultBusy=false;countyInput=null;
        if(render)showVaultUnlock();
    }
    private void showVaultUnlock(){
        page="locked";countyInput=null;root=column();root.setBackgroundColor(PAPER);padding(root,22,28,22,24);
        ScrollView lockedScroll=new ScrollView(this);LinearLayout c=column();lockedScroll.addView(c);add(root,lockedScroll);setContentView(root);
        title(c,"Path Home Private Vault");add(c,paragraph("Your private plan is locked. Enter your independent passphrase. Genyxis cannot reset it."));
        EditText pass=vaultPassword(c,"Vault passphrase");
        add(c,button("Unlock private plan",()->unlockVault(pass)));
        add(c,soft("Close application",()->finish()));
        if(!vaultNotice.isEmpty())add(c,paragraph(vaultNotice));
    }
    private SecretKey secret()throws Exception{KeyStore store=KeyStore.getInstance("AndroidKeyStore");store.load(null);if(store.containsAlias("reentry_local_v1"))return((KeyStore.SecretKeyEntry)store.getEntry("reentry_local_v1",null)).getSecretKey();
        KeyGenerator k=KeyGenerator.getInstance(KeyProperties.KEY_ALGORITHM_AES,"AndroidKeyStore");k.init(new KeyGenParameterSpec.Builder("reentry_local_v1",KeyProperties.PURPOSE_ENCRYPT|KeyProperties.PURPOSE_DECRYPT).setBlockModes(KeyProperties.BLOCK_MODE_GCM).setEncryptionPaddings(KeyProperties.ENCRYPTION_PADDING_NONE).setKeySize(256).build());return k.generateKey();}
    private String encrypt(String plaintext)throws Exception{Cipher c=Cipher.getInstance("AES/GCM/NoPadding");c.init(Cipher.ENCRYPT_MODE,secret());byte[] iv=c.getIV(),result=c.doFinal(plaintext.getBytes(StandardCharsets.UTF_8));return Base64.encodeToString(iv,Base64.NO_WRAP)+":"+Base64.encodeToString(result,Base64.NO_WRAP);}
    private String decrypt(String encrypted)throws Exception{String[] chunks=encrypted.split(":",2);Cipher c=Cipher.getInstance("AES/GCM/NoPadding");c.init(Cipher.DECRYPT_MODE,secret(),new GCMParameterSpec(128,Base64.decode(chunks[0],Base64.DEFAULT)));return new String(c.doFinal(Base64.decode(chunks[1],Base64.DEFAULT)),StandardCharsets.UTF_8);}
    private void put(String k,String v){if(vaultLocked||profileLocked)return;try{profile.put(k,v);queueSave();}catch(Exception e){toast("Could not update the local plan.");}}
    private String val(String k){return profile.optString(k,"");}
    private boolean yes(String k){return profile.optBoolean(k,false);}
    private void putBool(String k, boolean v){if(vaultLocked||profileLocked)return;try{profile.put(k,v);queueSave();}catch(Exception e){toast("Could not update the local plan.");}}

    private GradientDrawable bg(int color,int radius){GradientDrawable d=new GradientDrawable();d.setColor(color);d.setCornerRadius(dp(radius));return d;}
    private GradientDrawable stroke(int color,int edge){GradientDrawable d=bg(color,17);d.setStroke(dp(1),edge);return d;}
    private int dp(int n){return (int)(getResources().getDisplayMetrics().density*n+0.5f);}
    private LinearLayout column(){LinearLayout l=new LinearLayout(this);l.setOrientation(LinearLayout.VERTICAL);return l;}
    private LinearLayout row(){LinearLayout l=new LinearLayout(this);l.setOrientation(LinearLayout.HORIZONTAL);l.setGravity(Gravity.CENTER_VERTICAL);return l;}
    private void padding(View v,int l,int t,int r,int b){v.setPadding(dp(l),dp(t),dp(r),dp(b));}
    private void add(LinearLayout parent,View v){parent.addView(v,new LinearLayout.LayoutParams(-1,-2));}
    private TextView text(String s,int size,int color,boolean bold){TextView t=new TextView(this);t.setText(s);t.setTextSize(size);t.setTextColor(color);t.setGravity(Gravity.START|Gravity.CENTER_VERTICAL);if(bold)t.setTypeface(Typeface.DEFAULT,Typeface.BOLD);return t;}
    private TextView paragraph(String s){TextView t=text(s,16,MUTED,false);padding(t,0,5,0,13);return t;}
    private View spacer(int h){View v=new View(this);v.setLayoutParams(new LinearLayout.LayoutParams(1,dp(h)));return v;}
    private LinearLayout card(LinearLayout dest){LinearLayout c=column();c.setBackground(stroke(WHITE,0xffe2e6e3));padding(c,16,16,16,16);c.setElevation(dp(2));LinearLayout.LayoutParams p=new LinearLayout.LayoutParams(-1,-2);p.setMargins(dp(2),dp(7),dp(2),dp(9));dest.addView(c,p);return c;}
    private android.graphics.drawable.Drawable buttonSurface(boolean secondary){
        int fill=secondary?0xffedf6f3:TEAL;
        int edge=secondary?0xffa5c9bf:0xff09584e;
        GradientDrawable surface=bg(fill,14);
        surface.setStroke(dp(1),edge);
        return new android.graphics.drawable.RippleDrawable(
            android.content.res.ColorStateList.valueOf(secondary?0x220e7063:0x33ffffff),
            surface,bg(Color.WHITE,14));
    }
    private Button button(String s,Runnable action){
        Button b=new Button(this);
        b.setAllCaps(false);
        b.setText(s);
        b.setTextSize(16);
        b.setTypeface(Typeface.create("sans-serif-medium",Typeface.NORMAL));
        b.setLetterSpacing(0.01f);
        b.setTextColor(WHITE);
        b.setBackgroundTintList(null);
        b.setBackground(buttonSurface(false));
        b.setMinHeight(dp(52));
        b.setMinimumHeight(dp(52));
        b.setGravity(Gravity.CENTER);
        b.setPadding(dp(16),dp(12),dp(16),dp(12));
        b.setElevation(dp(2));
        b.setOnClickListener(v->action.run());
        return b;
    }
    private Button soft(String s,Runnable action){
        Button b=button(s,action);
        b.setTextColor(NAVY);
        b.setBackground(buttonSurface(true));
        b.setElevation(0);
        return b;
    }
    private void title(LinearLayout dest,String t){add(dest,text(t,25,NAVY,true));dest.addView(spacer(7));}
    private void small(LinearLayout dest,String t){add(dest,text(t,13,MUTED,false));}
    private ImageView photo(String image,int h){ImageView p=new ImageView(this);p.setImageResource(images.get(image));p.setScaleType(ImageView.ScaleType.CENTER_CROP);p.setBackground(bg(0xffd7e7df,14));p.setClipToOutline(true);p.setLayoutParams(new LinearLayout.LayoutParams(-1,dp(h)));return p;}
    private void show(String next){if(vaultLocked){showVaultUnlock();return;}flushSave();countyInput=null;page=next;root=column();root.setBackgroundColor(PAPER);setContentView(root);
        toolbar=row();toolbar.setBackgroundColor(NAVY);padding(toolbar,18,16,15,15);
        ImageView brandMark=new ImageView(this);brandMark.setImageResource(R.drawable.pathhome_brand);brandMark.setScaleType(ImageView.ScaleType.FIT_CENTER);toolbar.addView(brandMark,new LinearLayout.LayoutParams(dp(46),dp(46)));
        TextView app=text("PATH HOME",19,WHITE,true);padding(app,10,0,0,0);toolbar.addView(app,new LinearLayout.LayoutParams(0,dp(48),1));
        TextView privacyMenu=text("🔒",23,WHITE,true);privacyMenu.setContentDescription("Privacy and Private Vault");privacyMenu.setGravity(Gravity.CENTER);privacyMenu.setOnClickListener(v->show("privacy"));toolbar.addView(privacyMenu,new LinearLayout.LayoutParams(dp(46),dp(48)));
        TextView menu=text("⌂",28,WHITE,true);menu.setContentDescription("Go to Path Home start screen");menu.setGravity(Gravity.CENTER);menu.setOnClickListener(v->show("home"));toolbar.addView(menu,new LinearLayout.LayoutParams(dp(48),dp(48)));add(root,toolbar);
        scroll=new ScrollView(this);scroll.setFillViewport(false);content=column();padding(content,17,19,17,25);scroll.addView(content);root.addView(scroll,new LinearLayout.LayoutParams(-1,0,1));
        switch(next){case "home":home();break;case "plan":plan();break;case "timeline":timeline();break;case "allowed":allowed();break;case "allowed_results":results();break;case "money":money();break;case "help":help();break;case "appointments":appointments();break;case "family":family();break;case "progress":progress();break;case "unstuck":unstuck();break;case "privacy":privacy();break;case "stay":stayOnTrack();break;case "law":lawLibrary();break;case "mission":mission();break;default:home();}
        LinearLayout bottom=row();bottom.setBackgroundColor(NAVY);padding(bottom,4,6,4,8);
        for(String[] n:new String[][]{{"Home","home"},{"My Plan","plan"},{"Law Library","law"},{"Ask","allowed"},{"Help","help"}}){
            boolean selected=page.equals(n[1]);LinearLayout tab=column();tab.setGravity(Gravity.CENTER);
            tab.setBackground(bg(selected?0xff25655E:NAVY,11));
            ImageView glyph=new ImageView(this);glyph.setImageResource(iconForPage(n[1],selected));glyph.setScaleType(ImageView.ScaleType.FIT_CENTER);
            LinearLayout.LayoutParams gi=new LinearLayout.LayoutParams(dp(28),dp(28));gi.gravity=Gravity.CENTER;tab.addView(glyph,gi);
            TextView t=text(n[0],11,selected?0xffffd478:WHITE,true);t.setGravity(Gravity.CENTER);t.setMaxLines(2);
            padding(t,1,3,1,1);tab.addView(t,new LinearLayout.LayoutParams(-1,dp(31)));
            tab.setContentDescription("Open " + n[0]);tab.setOnClickListener(v->show(n[1]));
            LinearLayout.LayoutParams ti=new LinearLayout.LayoutParams(0,dp(69),1);ti.setMargins(dp(2),0,dp(2),0);bottom.addView(tab,ti);
        }add(root,bottom);
    }
    private int iconForPage(String target,boolean active){
        switch(target){
            case "home":return active?R.drawable.nav_home_active:R.drawable.nav_home_nav;
            case "plan": case "timeline":return active?R.drawable.nav_plan_active:R.drawable.nav_plan_nav;
            case "law":return active?R.drawable.nav_law_active:R.drawable.nav_law_nav;
            case "allowed":return active?R.drawable.nav_ask_active:R.drawable.nav_ask_nav;
            case "help":return active?R.drawable.nav_help_active:R.drawable.nav_help_nav;
            case "money":return active?R.drawable.nav_money_active:R.drawable.nav_money_nav;
            case "appointments":return active?R.drawable.nav_appointments_active:R.drawable.nav_appointments_nav;
            case "family":return active?R.drawable.nav_family_active:R.drawable.nav_family_nav;
            case "progress":return active?R.drawable.nav_progress_active:R.drawable.nav_progress_nav;
            case "unstuck":return active?R.drawable.nav_help_active:R.drawable.nav_help_nav;
            case "stay":return active?R.drawable.nav_stay_active:R.drawable.nav_stay_nav;
            default:return active?R.drawable.nav_home_active:R.drawable.nav_home_nav;
        }
    }
    private void hero(){LinearLayout c=card(content);add(c,photo("hero",172));c.addView(spacer(13));add(c,text("One clear path forward.",26,NAVY,true));add(c,paragraph("Get your next steps, important dates and help—without having to figure it all out alone."));}
    private void tile(String image,String name,String desc,String page){LinearLayout c=card(content);add(c,photo(image,111));c.addView(spacer(10));
        LinearLayout titleRow=row();ImageView glyph=new ImageView(this);glyph.setImageResource(iconForPage(page,false));
        glyph.setColorFilter(TEAL);titleRow.addView(glyph,new LinearLayout.LayoutParams(dp(32),dp(32)));
        TextView header=text(name,21,INK,true);padding(header,10,0,0,0);titleRow.addView(header,new LinearLayout.LayoutParams(0,-2,1));add(c,titleRow);
        add(c,paragraph(desc));add(c,soft("Open " + name + "  ›",()->show(page)));}
    private void home(){hero();
        LinearLayout welcome=card(content);title(welcome,"How can we help today?");add(welcome,paragraph("Pick one. You can change your mind at any time."));add(welcome,button("I'm coming home",()->show("plan")));welcome.addView(spacer(8));add(welcome,soft("I'm helping someone come home",()->show("family")));
        LinearLayout lawQuick=card(content);title(lawQuick,"Law Library — free case law");
        add(lawQuick,paragraph("Find court decisions and legal sources without a paid subscription."));
        add(lawQuick,button("Open Law Library",()->show("law")));
        if(!val("end").isEmpty()){LinearLayout c=card(content);title(c,"My supervision timeline");add(c,text(timelineSummary(),19,TEAL,true));add(c,soft("Review my dates",()->show("plan")));}
        LinearLayout reminderCard=card(content);
        title(reminderCard,"Stay on Track");
        add(reminderCard,paragraph("A private, optional reminder for places you choose. No location is sent to anyone."));
        add(reminderCard,button("Set up my private check-in",()->show("stay")));
        LinearLayout unblock=card(content);title(unblock,"Stuck on a step?");add(unblock,paragraph("Fees, paperwork, transportation or being told no? We can find another way forward."));add(unblock,button("I'm stuck — help me",()->show("unstuck")));
        tile("release_plan","My release plan","Release dates, first-day priorities and supervision needs.","plan");
        tile("appointments","Am I allowed to...?","Ask a question about travel, housing, work or your situation.","allowed");
        tile("money","My money & documents","Birth certificates, ID costs, and support if funds are short.","money");
        tile("local_help","Find help near me","Housing, food, documents, jobs and transportation.","help");
        tile("family","Help someone come home","Prepare a loved one's release before the big day.","family");
        tile("progress","My progress","Track the first things you've already completed.","progress");
        small(content,"Not a government service. Legal summaries are general information and may need review.");
        LinearLayout missionCard=card(content);title(missionCard,"Our Mission");add(missionCard,paragraph("Because freedom deserves a fighting chance."));add(missionCard,soft("Read the full mission statement  ›",()->show("mission")));
        add(content,soft("Privacy & Private Vault",()->show("privacy")));
        small(content,"Published by Genyxis Network");
    }
    private EditText input(LinearLayout c,String label,String key,String hint){add(c,text(label,16,INK,true));EditText e=new EditText(this);if("county".equals(key))countyInput=e;e.setSingleLine(true);e.setTextSize(18);e.setHint(hint);e.setText(val(key));e.setBackground(stroke(0xfffafcfb,0xffc8d6d4));padding(e,14,11,12,11);LinearLayout.LayoutParams p=new LinearLayout.LayoutParams(-1,dp(50));p.setMargins(0,dp(7),0,dp(16));c.addView(e,p);e.addTextChangedListener(new android.text.TextWatcher(){public void beforeTextChanged(CharSequence s,int st,int count,int after){}public void onTextChanged(CharSequence s,int st,int before,int count){put(key,s.toString());}public void afterTextChanged(android.text.Editable a){}});return e;}
    private void pickDate(LinearLayout c,String label,String key){add(c,text(label,16,INK,true));String initial=val(key);TextView choice=text(initial.isEmpty()?"Choose a date (or leave unknown)":initial,18,TEAL,true);padding(choice,14,15,14,15);choice.setBackground(stroke(0xffedf6f3,0xffbddbd3));LinearLayout.LayoutParams p=new LinearLayout.LayoutParams(-1,-2);p.setMargins(0,dp(6),0,dp(9));c.addView(choice,p);
        choice.setOnClickListener(v->{Calendar cal=Calendar.getInstance();String lastChosen=val(key);if(lastChosen.matches("\\d{4}-\\d{2}-\\d{2}")){try{Date d=new SimpleDateFormat("yyyy-MM-dd",Locale.US).parse(lastChosen);cal.setTime(d);}catch(Exception ignored){}}new DatePickerDialog(this,(view,y,m,d)->{String iso=String.format(Locale.US,"%04d-%02d-%02d",y,m+1,d);put(key,iso);choice.setText(iso);},cal.get(Calendar.YEAR),cal.get(Calendar.MONTH),cal.get(Calendar.DAY_OF_MONTH)).show();});
        add(c,soft("I don't know this date yet",()->{put(key,"");choice.setText("Choose a date (or leave unknown)");}));c.addView(spacer(14));}
    private void stateChooser(LinearLayout c){add(c,text("Where will you live after release?",17,INK,true));Spinner sp=new Spinner(this);ArrayAdapter<String> adapter=new ArrayAdapter<>(this,android.R.layout.simple_spinner_dropdown_item,jurisdictions);sp.setAdapter(adapter);int selection=stateCodes.indexOf(val("state"));if(selection>=0)sp.setSelection(selection);c.addView(sp,new LinearLayout.LayoutParams(-1,dp(55)));sp.setOnItemSelectedListener(new android.widget.AdapterView.OnItemSelectedListener(){public void onNothingSelected(android.widget.AdapterView<?> a){}public void onItemSelected(android.widget.AdapterView<?> a,View v,int position,long id){if(position<0||position>=stateCodes.size())return;String chosen=stateCodes.get(position);if(!chosen.equals(val("state"))){put("state",chosen);put("county","");if(countyInput!=null)countyInput.setText("");toast("State updated. Recheck your county for the right local services.");}}});}
    private void plan(){title(content,"My personal release plan");add(content,paragraph("Five short steps. Start with location, then supervision dates, appointments and help needs."));
        LinearLayout c=card(content);title(c,"Step 1 of 5 · My location");stateChooser(c);input(c,"County (optional)","county","Example: Jackson");input(c,"ZIP code (optional)","zip","Example: 49201");
        add(c,text("What kind of supervision applies?",17,INK,true));
        String[] kinds={"I don't know yet","State parole","State probation","Federal supervised release","No supervision"};
        Spinner type=new Spinner(this);type.setAdapter(new ArrayAdapter<>(this,android.R.layout.simple_spinner_dropdown_item,kinds));
        int selectedType=0;for(int i=1;i<kinds.length;i++)if(kinds[i].equals(val("supervision")))selectedType=i;
        type.setSelection(selectedType);c.addView(type,new LinearLayout.LayoutParams(-1,dp(55)));
        type.setOnItemSelectedListener(new android.widget.AdapterView.OnItemSelectedListener(){
            public void onNothingSelected(android.widget.AdapterView<?> a){}
            public void onItemSelected(android.widget.AdapterView<?> a,View v,int pos,long id){String picked=pos==0?"":kinds[pos];if(!picked.equals(val("supervision")))put("supervision",picked);}
        });
        add(c,soft("Next: supervision dates  ›",()->show("timeline")));
        LinearLayout n=card(content);title(n,"First things to take care of");for(String s:new String[]{"Confirm when and where you must report","Birth certificate, Social Security card and state ID","Housing and transportation","Food, healthcare and phone access"})add(n,paragraph("• "+s));add(n,soft("Track these steps",()->show("progress")));
    }
    private void timeline(){title(content,"How long is my supervision?");add(content,paragraph("Your dates shape the plan. The expected end is not official discharge."));
        LinearLayout c=card(content);title(c,"Step 2 of 5 · Important dates");pickDate(c,"Release date","release");pickDate(c,"Supervision start date","start");pickDate(c,"Expected supervision end date","end");
        add(c,text("How long would you like support?",16,INK,true));
        for(String duration:new String[]{"First 90 days","6 months","12 months","Through supervision","Ongoing"}){RadioButton rb=new RadioButton(this);rb.setText(duration);rb.setTextSize(17);rb.setChecked(duration.equals(val("support")));c.addView(rb);rb.setOnClickListener(v->{put("support",duration);show("timeline");});}
        if(!val("start").isEmpty()&&!val("end").isEmpty()&&val("start").compareTo(val("end"))>0){add(c,text("Please check your dates. The end appears to be before the start.",17,0xffad3b16,true));}
        add(c,button("Next: appointments  ›",()->{if(!WorkflowRules.datesInOrder(val("start"),val("end"))){toast("End date cannot come before the start date. Please correct it.");return;}show("appointments");}));
        LinearLayout l=card(content);title(l,"Time on supervision");add(l,paragraph(timelineSummary()));
    }
    private String timelineSummary(){String e=val("end");if(e.isEmpty())return "No end date entered. You can still use all help tools.";try{SimpleDateFormat f=new SimpleDateFormat("yyyy-MM-dd",Locale.US);f.setLenient(false);Date d=f.parse(e);int days=(int)Math.ceil((d.getTime()-System.currentTimeMillis())/86400000.0);if(days>=0)return "Expected end: "+e+" · About "+days+" days left. Confirm with your supervising agency.";return "The expected end date has passed. Confirm official discharge before assuming supervision is over.";}catch(Exception x){return "Please check the expected end date.";}}
    private void appointments(){title(content,"Appointments and check-ins");add(content,paragraph("Only use dates from your paperwork or confirmed appointments. No automatic background reminders in this version."));
        LinearLayout c=card(content);title(c,"Step 3 of 5 · Next required appointment");pickDate(c,"Next required reporting date","appointmentDate");input(c,"Confirmed appointment time (HH:mm, 24-hour)","appointmentTime","Example: 09:00");input(c,"Appointment details / location","appointmentText","Example: office address");add(c,button("Add this to my phone calendar",()->calendar()));
        c.addView(spacer(7));add(c,soft("Next: money and documents",()->show("money")));
    }
    private void calendar(){
        String d=val("appointmentDate"),t=val("appointmentTime").trim();
        if(d.isEmpty()){toast("Choose the confirmed appointment date first.");return;}
        if(!t.matches("(?:[01]\\d|2[0-3]):[0-5]\\d")){toast("Enter the confirmed local appointment time, for example 09:00. Never guess a reporting time.");return;}
        try{
            SimpleDateFormat f=new SimpleDateFormat("yyyy-MM-dd HH:mm",Locale.US);
            f.setLenient(false);
            Date date=f.parse(d+" "+t);
            Intent i=new Intent(Intent.ACTION_INSERT,CalendarContract.Events.CONTENT_URI);
            i.putExtra(CalendarContract.EXTRA_EVENT_BEGIN_TIME,date.getTime());
            i.putExtra(CalendarContract.EXTRA_EVENT_END_TIME,date.getTime()+3600000L);
            i.putExtra(CalendarContract.Events.TITLE,"Important appointment — confirm details");
            i.putExtra(CalendarContract.Events.DESCRIPTION,val("appointmentText")+" — Confirm all details against the original notice.");
            startActivity(i);
        }catch(Exception e){toast("Calendar unavailable or date/time invalid. Recheck the notice.");}
    }
    private void money(){title(content,"Money for ID and first steps");add(content,paragraph("A replacement Social Security card is free. Check current birth-record and ID fees with the issuing agency."));
        LinearLayout c=card(content);title(c,"Step 4 of 5 · My budget");input(c,"Money available ($)","cash","0");input(c,"Birth certificate estimate ($)","birthCost","Check fee");input(c,"State ID or license estimate ($)","idCost","Check fee");input(c,"Travel, postage and other costs ($)","otherCost","0");
        boolean valid=WorkflowRules.validMoney(val("cash"))&&WorkflowRules.validMoney(val("birthCost"))&&WorkflowRules.validMoney(val("idCost"))&&WorkflowRules.validMoney(val("otherCost"));
        if(!valid)add(c,text("Check your amounts: use numbers of $0 or more.",18,0xffa14c1b,true));
        boolean unknown=val("birthCost").trim().isEmpty()||val("idCost").trim().isEmpty();
        double need=Math.max(0,amount("birthCost")+amount("idCost")+amount("otherCost")-amount("cash"));
        if(valid)add(c,text(String.format(Locale.US,"%s: $%.2f",unknown?"Partial shortfall (fees missing)":"Estimated shortfall",need),24,TEAL,true));add(c,paragraph("If a required fee is unknown, this total is incomplete. Ask about waivers and alternative funding."));add(c,soft("Update total",()->show("money")));add(c,button("Find assistance near me  ›",()->show("help")));
        LinearLayout d=card(content);title(d,"Your first papers");for(String s:new String[]{"Birth certificate","Social Security card (replacement is free)","State ID or driver license"})add(d,paragraph("• "+s));
    }
    private double amount(String k){try{double v=Double.parseDouble(val(k).replace("$","").trim());return Double.isNaN(v)||Double.isInfinite(v)?0:Math.max(0,v);}catch(Exception e){return 0;}}
    private void allowed(){title(content,"Am I Allowed To...?");add(content,paragraph("Ask about travel, work, driving, housing, relationships, voting or a situation that doesn't fit a standard category."));
        LinearLayout c=card(content);input(c,"What do you need to know?","question","Example: Can I move to another county?");
        add(c,button("Search legal information",()->{if(val("question").trim().isEmpty()){toast("Type a question first.");return;}show("allowed_results");}));
        for(String q:new String[]{"Can I travel outside my county?","Can I live with family?","Can I get a job?","Can I vote?","Can I move to a new address?"})add(c,soft(q,()->{put("question",q);show("allowed_results");}));
        LinearLayout warn=card(content);add(warn,text("Important: a general law is not individual permission.",17,0xffa14c1b,true));add(warn,paragraph("Your signed order and instructions from the supervising agency control your case. Ask your officer or attorney when uncertain."));
    }
    private void results(){
        title(content,"Research for your question");
        String code=val("state").trim().toUpperCase(Locale.US);
        String q=val("question").trim();
        LinearLayout c=card(content);
        if(q.isEmpty()){add(c,paragraph("Type a question to begin."));add(c,button("Ask a question",()->show("allowed")));return;}
        add(c,text("You asked: "+q,18,NAVY,true));c.addView(spacer(10));
        if(code.isEmpty()){
            add(c,paragraph("Choose your state first. Different states have different rules, and Path Home will not guess."));
            add(c,button("Select my state",()->show("plan")));
            add(c,soft("Open the free Law Library",()->show("law")));return;
        }
        String topic=WorkflowRules.questionTopic(q);
        if("multiple".equals(topic)){
            add(c,paragraph("Your situation touches more than one legal topic. Ask about each part separately—for example, where you may live AND whether travel approval is required. Confirm your specific conditions with your officer or attorney."));
            add(c,soft("Ask a more specific question",()->show("allowed")));
            add(c,soft("Open free Law Library",()->show("law")));return;
        }
        if(topic.isEmpty()){
            add(c,paragraph("We do not have a verified rule matching this particular situation. Do not assume it is allowed or prohibited."));
            add(c,button("Open free Law Library",()->show("law")));
            add(c,soft("Find local legal help",()->{helpCategory="legal";show("help");}));
            add(c,soft("Ask a different question",()->show("allowed")));return;
        }
        JSONArray rules=legal.optJSONArray("legal_rules");int count=0;
        if(rules!=null)for(int i=0;i<rules.length()&&count<4;i++){
            JSONObject r=rules.optJSONObject(i);if(r==null)continue;
            String state=r.optString("jurisdiction_code"),rt=r.optString("topic_code");
            if(!state.equals(code)&&!("FED".equals(state)&&"Federal supervised release".equals(val("supervision"))))continue;
            if(!rt.contains(topic))continue;
            count++;
            add(c,text(r.optString("rule_title")+" ["+state+"]",18,NAVY,true));
            add(c,paragraph(r.optString("plain_language")));
            add(c,text("General information · Confirm your personal conditions · Review: "+r.optString("legal_review_status"),13,0xffa14c1b,true));
            String url=r.optString("source_url");if(url.startsWith("https://"))add(c,soft("Read cited source ↗",()->visit(url)));
            c.addView(spacer(12));
        }
        if(count==0)add(c,paragraph("No matching source-checked rule is in this copy. Do not assume permission; use the Law Library or consult an attorney or supervising office."));
        add(c,soft("Ask a different question",()->show("allowed")));
        add(c,soft("Open Law Library",()->show("law")));
    }

    private void visit(String url){try{if(url.startsWith("https://"))startActivity(new Intent(Intent.ACTION_VIEW,Uri.parse(url)));}catch(Exception e){toast("Could not open the source link.");}}
        // A person's chosen state/county are the ONLY inputs to local referral matching.
    // This service never requests GPS or sends location to Genyxis.
    private String helpCategory="All";
    private String helpFromBarrier(String barrier){
      switch(barrier){
        case "Money or fees": return "money";
        case "Missing ID or paperwork":return "documents";
        case "No transportation":return "transportation";
        case "Need safe housing":return "housing";
        case "Instructions unclear":case "Program said no":return "legal";
        default:return "All";
      }
    }
    private void localHelpForBarrier(String barrier){helpCategory=helpFromBarrier(barrier);show("help");}
    private void help(){
       title(content,"Find real help near home");
       small(content,"Step 5 of 5 · Find a real local contact. You can return here anytime.");
       String code=val("state").trim().toUpperCase(Locale.US),county=val("county").trim();
       add(content,paragraph("Showing services for " + (county.isEmpty()?"county not set":county) + ", " + (code.isEmpty()?"state not set":code) + ". This is the location YOU entered, not GPS tracking."));
       if(code.isEmpty()||county.isEmpty()){
         LinearLayout no=card(content);
         add(no,text("Choose where you'll be living",19,NAVY,true));
         add(no,paragraph("We will not guess which office serves you. Set your state and county in My Plan first."));
         add(no,button("Choose my state and county",()->show("plan")));
         add(no,soft("Find local assistance through 211",()->visit("https://211.org/")));
         add(no,soft("Review my first steps",()->show("progress")));
         return;
       }
       LinearLayout categories=card(content);title(categories,"What do you need right now?");
       for(String category:new String[]{"All","documents","housing","employment","transportation","legal","money"}){
         final String use=category;
         add(categories,soft((helpCategory.equals(category)?"✓ ":"")+category.toUpperCase(Locale.US),()->{helpCategory=use;show("help");}));
       }
       JSONArray listing=legal.optJSONArray("local_providers");int count=0;
       if(listing!=null)for(int i=0;i<listing.length();i++){
          JSONObject p=listing.optJSONObject(i);if(p==null)continue;
          if(!code.equalsIgnoreCase(p.optString("state_code"))||!county.replaceAll("(?i)\\s+county$","").trim().equalsIgnoreCase(p.optString("county")))continue;
          JSONArray cats=p.optJSONArray("categories");boolean match=helpCategory.equals("All");
          if(cats!=null)for(int j=0;j<cats.length();j++)if(helpCategory.equalsIgnoreCase(cats.optString(j)))match=true;
          if(!match||!p.optString("source_url").startsWith("https://")||p.optString("verified_on").isEmpty())continue;
          count++;
          LinearLayout c=card(content);
          add(c,text(p.optString("name"),20,NAVY,true));
          add(c,paragraph("Helps with: "+p.optString("scope")));
          if(!p.optString("address").isEmpty())add(c,paragraph("Address: "+p.optString("address")));
          add(c,text("Phone: "+p.optString("phone"),17,TEAL,true));
          add(c,paragraph("Before you go: "+p.optString("before_you_go")));
          add(c,paragraph("Eligibility: "+p.optString("eligibility")));
          add(c,paragraph("If turned away: "+p.optString("if_denied")));
          add(c,text("Official source checked: "+p.optString("verified_on")+" · Confirm current availability",12,MUTED,false));
          String url=p.optString("website");if(url.startsWith("https://"))add(c,soft("Official website / instructions  ↗",()->visit(url)));
          String phone=p.optString("phone").replaceAll("[^0-9]","");
          if(phone.length()>=10)add(c,soft("Call this organization  ↗",()->{try{startActivity(new Intent(Intent.ACTION_DIAL,Uri.parse("tel:"+phone)));}catch(Exception ignored){toast("Phone dialer not available.");}}));
       }
       if(count==0){LinearLayout c=card(content);add(c,text("No verified local match yet",19,NAVY,true));add(c,paragraph("We won't send you to a random office. Use the local 211 network to find a provider covering your county."));}
       LinearLayout backup=card(content);title(backup,"Before you travel");
       add(backup,paragraph("Call first. Confirm service area, hours, eligibility, documents, cost, and whether an appointment is necessary."));
       add(backup,button("Find local alternatives through 211",()->visit("https://211.org/")));
       add(backup,soft("I was turned away — help me",()->show("unstuck")));
       add(backup,soft("Change my county",()->show("plan")));
       add(backup,button("Review my first steps  ›",()->show("progress")));
       add(backup,paragraph("A nearby organization is not automatically your assigned reporting office. Follow your actual supervision instructions."));
    }
    private void family(){title(content,"Help someone come home");add(content,paragraph("A parent, spouse, sibling or friend can help plan before release. The returning person should choose what they share."));
        LinearLayout c=card(content);add(c,photo("family",148));c.addView(spacer(12));title(c,"Family preparation checklist");for(String t:FAMILY_TASKS){CheckBox cb=new CheckBox(this);cb.setText(t);cb.setTextSize(17);cb.setChecked(yes("family_"+t));padding(cb,4,7,4,7);c.addView(cb);cb.setOnCheckedChangeListener((b,v)->putBool("family_"+t,v));}
        LinearLayout note=card(content);add(note,text("Private handoff is not live yet",17,0xff9c521f,true));add(note,paragraph("This offline app cannot invite or register someone else remotely. A verified account service and consent workflow are required before live sharing."));add(note,soft("See the returning person's plan",()->show("plan")));
    }
    private void progress(){title(content,"Your first steps");add(content,paragraph("It's okay to do one thing at a time. Check off what you've handled. If a step is blocked, we will find a different route."));
        LinearLayout c=card(content);int done=0;for(String s:START_TASKS)if(yes("done_"+s))done++;add(c,text(done+" of "+START_TASKS.length+" completed",23,TEAL,true));for(String t:START_TASKS){CheckBox cb=new CheckBox(this);cb.setText(t);cb.setTextSize(17);cb.setChecked(yes("done_"+t));padding(cb,4,8,4,8);c.addView(cb);cb.setOnCheckedChangeListener((b,v)->putBool("done_"+t,v));}add(c,soft("Refresh my progress",()->show("progress")));add(c,button("A step is blocked? Get help",()->show("unstuck")));
        LinearLayout d=card(content);title(d,"My timeline");add(d,paragraph(timelineSummary()));add(d,soft("Update supervision dates",()->show("timeline")));
    }
    /** Optional mindful check-in, stored locally within the Keystore-encrypted plan. */

    // Lean improvement for the *process*: identify an obstacle, suggest a single manageable next step.
    // The person is never scored or automatically reported. Answers stay in their encrypted local profile.
    private String barrierAction(String barrier){
        switch(barrier){
            case "Money or fees":return "Ask the issuing agency about a fee waiver. If none is available, ask 211 or a reentry service about help with document costs.";
            case "Missing ID or paperwork":return "Ask the agency what alternative proof it accepts. Start with birth records, then Social Security, then state ID.";
            case "No transportation":return "Confirm a bus route or ride with local transit or 211. If you may miss required reporting, contact your supervising office promptly.";
            case "Program said no":return "Ask why you were not eligible, whether you can appeal, and what other proof they accept. Ask for a referral.";
            case "Instructions unclear":return "Check your signed paperwork and ask the issuing agency, an attorney or legal aid to clarify before acting.";
            case "Need safe housing":return "Ask 211 or local reentry housing providers about options. Confirm any proposed address with your supervising office if required.";
            case "Something else":return "Write down the problem in one sentence and find one person or agency that can answer it. Ask for a referral if they cannot help.";
            default:return "Choose what is blocking you. We will suggest a small step, not judge your progress.";
        }
    }
    private void unstuck(){
        title(content,"I'm stuck — help me find a way");
        add(content,paragraph("Sometimes the process is the problem, not you. Pick what is blocking your next step. This stays on your phone."));
        LinearLayout goal=card(content);input(goal,"What are you trying to get done? (optional)","unstuck_goal","Example: Get a birth certificate");
        LinearLayout problems=card(content);title(problems,"What's getting in the way?");
        String[] barriers={"Money or fees","Missing ID or paperwork","No transportation","Program said no","Instructions unclear","Need safe housing","Something else"};
        for(String b:barriers){add(problems,soft((b.equals(val("unstuck_barrier"))?"✓  ":"")+b,()->{put("unstuck_barrier",b);show("unstuck");}));problems.addView(spacer(5));}
        LinearLayout next=card(content);title(next,"One next step");add(next,paragraph(barrierAction(val("unstuck_barrier"))));
        if(!val("unstuck_barrier").isEmpty()){
            add(next,text("My chosen obstacle: "+val("unstuck_barrier"),16,TEAL,true));
            add(next,paragraph("If this suggestion does not fit, choose a different problem or ask a local service for a referral."));
        }
        add(next,button("Find a local service for this exact problem",()->localHelpForBarrier(val("unstuck_barrier"))));
        next.addView(spacer(6));add(next,soft("Back to my progress",()->show("progress")));
        add(content,paragraph("This does not change parole rules, reporting deadlines, or agency eligibility. Confirm requirements with the responsible office or an attorney."));
    }
    private void lawLibrary(){
        title(content,"Law Library");
        add(content,paragraph("Free legal research—one place to start. Court decisions can be old, from another state, or overturned. Always check the court, date, later rulings, and the conditions on your own release documents."));
        LinearLayout sources=card(content);
        title(sources,"Find court decisions (case law)");
        add(sources,button("Search free case law — CourtListener",()->visit("https://www.courtlistener.com/")));
        sources.addView(spacer(8));
        add(sources,soft("Find cases — Google Scholar",()->visit("https://scholar.google.com/")));
        sources.addView(spacer(8));
        add(sources,soft("Free legal research guide — Library of Congress",()->visit("https://guides.loc.gov/free-case-law")));
        add(sources,paragraph("These links open public resources online. Their coverage and update schedules differ; case-law searching is not yet an offline, indexed database in this app."));
        LinearLayout state=card(content);
        title(state,"Laws and sources for my state");
        String code=val("state").toUpperCase(Locale.US);
        add(state,paragraph(code.isEmpty()?"Choose your state in My Plan first to see relevant source links.":"Showing available research for "+code+". This is a research directory, not an attorney-approved statement of your conditions."));
        JSONArray rules=legal.optJSONArray("legal_rules");int count=0;
        if(rules!=null)for(int i=0;i<rules.length()&&count<12;i++){
            JSONObject obj=rules.optJSONObject(i);if(obj==null)continue;
            if(!obj.optString("jurisdiction_code").equals(code)&&!("FED".equals(obj.optString("jurisdiction_code"))&&"Federal supervised release".equals(val("supervision"))))continue;
            count++;
            add(state,text(obj.optString("rule_title")+" ["+obj.optString("jurisdiction_code")+"]",17,NAVY,true));
            add(state,paragraph(obj.optString("plain_language")));
            String url=obj.optString("source_url");
            if(url.startsWith("https://"))add(state,soft("Read source ↗",()->visit(url)));
            state.addView(spacer(10));
        }
        if(count==0)add(state,paragraph("No source-checked entry for this state is available in the local starter records. Use the free case-law resources above and seek qualified legal help for case-specific questions."));
        add(state,soft("Change my state",()->show("plan")));
    }
    private void stayOnTrack(){
        title(content,"Stay on Track");
        add(content,paragraph("Choose your own reminder places. When you open Path Home or tap Check now, your phone can check whether you're near one of them. This is not legal supervision, and it never tells anyone where you are."));
        LinearLayout choice=card(content);
        title(choice,"Your choice, your privacy");
        CheckBox active=new CheckBox(this);
        active.setText("Enable private location check-ins"); active.setTextSize(18);
        active.setChecked(yes("stay_enabled"));
        padding(active,0,10,0,14);add(choice,active);
        active.setOnCheckedChangeListener((btn,on)->putBool("stay_enabled",on));
        add(choice,paragraph("Default: OFF. Location stays on your phone. Nothing is reported to family, supervision agencies, or Path Home. If disabled, automatic on-open checks stop immediately."));
        CheckBox vibration=new CheckBox(this);
        vibration.setText("Gently vibrate when near a reminder place");
        vibration.setTextSize(17);vibration.setChecked(yes("stay_vibrate"));add(choice,vibration);
        vibration.setOnCheckedChangeListener((btn,on)->putBool("stay_vibrate",on));
        LinearLayout zones=card(content);
        title(zones,"Places I want a reminder");
        add(zones,paragraph("Add a place while you're physically there. No one else will see the label or its location. You choose the places and can delete them anytime."));
        input(zones,"A private name for this place","stay_new_label","Example: Old hangout");
        add(zones,text("Remind me within approximately:",16,INK,true));
        String[] sizes={"150 meters","300 meters","500 meters","1 kilometer"};
        Spinner radius=new Spinner(this);
        radius.setAdapter(new ArrayAdapter<>(this,android.R.layout.simple_spinner_dropdown_item,sizes));
        int savedRadius=(int)profile.optDouble("stay_radius",300);
        radius.setSelection(savedRadius==150?0:savedRadius==500?2:savedRadius==1000?3:1);
        add(zones,radius);
        radius.setOnItemSelectedListener(new android.widget.AdapterView.OnItemSelectedListener(){
           public void onNothingSelected(android.widget.AdapterView<?> a){}
           public void onItemSelected(android.widget.AdapterView<?> a,View v,int pos,long id){
               try{profile.put("stay_radius",new int[]{150,300,500,1000}[pos]);saveProfile();}catch(Exception ignored){}
           }
        });
        zones.addView(spacer(12));
        add(zones,button("Save my current place as a reminder",()->{
            if(!yes("stay_enabled")){toast("Turn on private check-ins first.");return;}
            if(val("stay_new_label").trim().isEmpty()){toast("Give the place a private label first.");return;}
            awaitingSaveLocation=true;manualCheck=false;requestLocalLocation();
        }));
        zones.addView(spacer(10));
        add(zones,soft("Check my location now",()->{
            if(!yes("stay_enabled")){toast("Turn on private check-ins first.");return;}
            awaitingSaveLocation=false;manualCheck=true;requestLocalLocation();
        }));
        JSONArray a=profile.optJSONArray("stay_zones");
        LinearLayout saved=card(content);title(saved,"My saved reminder places");
        if(a==null||a.length()==0)add(saved,paragraph("No places saved yet. You can still use the other Path Home features."));
        else for(int i=0;i<a.length();i++){
            JSONObject obj=a.optJSONObject(i);if(obj==null)continue;
            final int index=i;
            String name=obj.optString("label","My reminder place");
            add(saved,paragraph("• "+name+"  (about "+obj.optInt("radius",300)+" m)"));
            add(saved,soft("Remove: "+name,()->{
                JSONArray current=profile.optJSONArray("stay_zones"),updated=new JSONArray();
                if(current!=null)for(int k=0;k<current.length();k++)if(k!=index)updated.put(current.opt(k));
                try{profile.put("stay_zones",updated);saveProfile();show("stay");}catch(Exception ignored){}
            }));
        }
        LinearLayout guard=card(content);
        title(guard,"You stay in control");
        add(guard,paragraph("This check cannot determine whether a place is legally permitted or safe. Check your signed conditions and follow your official instructions. The app does not identify or label any neighborhood as dangerous."));
        add(guard,paragraph("On-open checks and manual checks are implemented here. Continuous background geofence alerts are not yet enabled and need separate permission and device testing."));
        add(guard,soft("Turn off and remove all reminder places",()->new android.app.AlertDialog.Builder(this)
            .setTitle("Remove private check-ins?")
            .setMessage("This disables check-ins and deletes all saved reminder places from this phone.")
            .setNegativeButton("Keep",null)
            .setPositiveButton("Remove",(d,w)->{try{profile.remove("stay_zones");profile.remove("stay_last_alert");profile.put("stay_enabled",false);saveProfile();show("stay");}catch(Exception ignored){}}).show()));
    }
    private boolean hasForegroundLocation(){
        return checkSelfPermission(Manifest.permission.ACCESS_FINE_LOCATION)==PackageManager.PERMISSION_GRANTED
            || checkSelfPermission(Manifest.permission.ACCESS_COARSE_LOCATION)==PackageManager.PERMISSION_GRANTED;
    }
    private void requestLocalLocation(){
        if(!hasForegroundLocation()){
            requestPermissions(new String[]{Manifest.permission.ACCESS_FINE_LOCATION,Manifest.permission.ACCESS_COARSE_LOCATION},LOCATION_REQUEST);
            return;
        }
        fetchLocalLocation();
    }
    @Override public void onRequestPermissionsResult(int requestCode,String[] permissions,int[] grantResults){
        super.onRequestPermissionsResult(requestCode,permissions,grantResults);
        if(requestCode==LOCATION_REQUEST){
            if(hasForegroundLocation())fetchLocalLocation();
            else { awaitingSaveLocation=false;manualCheck=false;toast("Location permission was not granted. No check was made."); }
        }
    }
    @SuppressWarnings("MissingPermission")
    private void fetchLocalLocation(){
        if(vaultLocked||profileLocked||!foreground||!hasForegroundLocation() || checkInProgress)return;
        final LocationManager lm=(LocationManager)getSystemService(Context.LOCATION_SERVICE);
        if(lm==null){toast("Location is unavailable on this phone.");return;}
        String provider=null;
        if(lm.isProviderEnabled(LocationManager.GPS_PROVIDER) &&
            checkSelfPermission(Manifest.permission.ACCESS_FINE_LOCATION)==PackageManager.PERMISSION_GRANTED)
            provider=LocationManager.GPS_PROVIDER;
        else if(lm.isProviderEnabled(LocationManager.NETWORK_PROVIDER))provider=LocationManager.NETWORK_PROVIDER;
        if(provider==null){toast("Turn on location services to check this place.");return;}
        checkInProgress=true;
        final int requestGeneration=vaultGeneration;
        try{
            if(Build.VERSION.SDK_INT>=30){
                lm.getCurrentLocation(provider,new CancellationSignal(),getMainExecutor(),loc->{if(requestGeneration!=vaultGeneration)return;checkInProgress=false;onLocalLocation(loc);});
            }else{
                lm.requestSingleUpdate(provider,new LocationListener(){
                    @Override public void onLocationChanged(Location loc){if(requestGeneration!=vaultGeneration)return;checkInProgress=false;onLocalLocation(loc);}
                    @Override public void onProviderDisabled(String p){if(requestGeneration!=vaultGeneration)return;checkInProgress=false;toast("Location provider is off.");}
                },null);
            }
        }catch(Exception ex){checkInProgress=false;toast("Could not check location. Please try again.");}
    }
    private void onLocalLocation(Location here){
        if(vaultLocked||profileLocked||!foreground)return;
        if(here==null){awaitingSaveLocation=false;manualCheck=false;toast("Could not get a current location. Try outdoors or near a window.");return;}
        if(!yes("stay_enabled")){manualCheck=false;awaitingSaveLocation=false;toast("Private check-ins are turned off.");return;}
        if(awaitingSaveLocation){
            awaitingSaveLocation=false;
            int meters=(int)profile.optDouble("stay_radius",300);
            if(here.hasAccuracy() && here.getAccuracy()>meters){toast("Location isn't accurate enough. Try again outside.");return;}
            JSONArray zones=profile.optJSONArray("stay_zones");if(zones==null)zones=new JSONArray();
            if(zones.length()>=20){toast("Remove a saved place before adding another.");return;}
            try{
                JSONObject zone=new JSONObject();zone.put("label",val("stay_new_label").trim());
                zone.put("lat",here.getLatitude());zone.put("lon",here.getLongitude());zone.put("radius",meters);
                zones.put(zone);profile.put("stay_zones",zones);profile.remove("stay_new_label");saveProfile();
                toast("Saved only on this phone.");show("stay");
            }catch(Exception ex){toast("Could not save place.");}
            return;
        }
        final boolean userRequested=manualCheck;manualCheck=false;
        JSONArray zones=profile.optJSONArray("stay_zones");
        if(zones==null||zones.length()==0){if(userRequested)toast("No reminder places saved. Nothing was checked.");return;}
        for(int i=0;i<zones.length();i++){
            JSONObject z=zones.optJSONObject(i);if(z==null)continue;
            float[] distance=new float[1];
            Location.distanceBetween(here.getLatitude(),here.getLongitude(),z.optDouble("lat"),z.optDouble("lon"),distance);
            int radius=z.optInt("radius",300);
            if(here.hasAccuracy() && here.getAccuracy()>radius)continue;
            if(distance[0]<=radius){
                long now=System.currentTimeMillis(),last=profile.optLong("stay_last_alert",0);
                if(now-last<REMINDER_COOLDOWN_MS){if(userRequested)toast("You've recently received a reminder. No new alert needed yet.");return;}
                try{profile.put("stay_last_alert",now);saveProfile();}catch(Exception ignored){}
                if(yes("stay_vibrate")){
                    Vibrator vibrator=(Vibrator)getSystemService(Context.VIBRATOR_SERVICE);
                    if(vibrator!=null&&vibrator.hasVibrator())vibrator.vibrate(VibrationEffect.createOneShot(240,VibrationEffect.DEFAULT_AMPLITUDE));
                }
                new android.app.AlertDialog.Builder(this).setTitle(REMINDER_PROMPT)
                    .setMessage(REMINDER_SUPPORT)
                    .setPositiveButton("Thanks. I've got this.",null).show();
                return;
            }
        }
        if(userRequested)toast("Check complete. No personal reminder places nearby. Your current location was not saved.");
    }
    @Override protected void onResume(){
        super.onResume();
        if(vaultLocked||profileLocked)return;
        if(!awaitingSaveLocation && yes("stay_enabled") && profile.optJSONArray("stay_zones")!=null &&
           profile.optJSONArray("stay_zones").length()>0 && hasForegroundLocation()){
            awaitingSaveLocation=false;manualCheck=false;fetchLocalLocation();
        }
    }
    private void mission(){
        title(content,"Our Mission");
        add(content,text("Because freedom deserves a fighting chance.",19,TEAL,true));
        content.addView(spacer(10));
        try(InputStream f=getAssets().open("mission.txt")){
            ByteArrayOutputStream b=new ByteArrayOutputStream();byte[] buf=new byte[4096];int n;
            while((n=f.read(buf))!=-1)b.write(buf,0,n);
            String[] sections=b.toString(StandardCharsets.UTF_8.name()).trim().split("\\n\\s*\\n");
            for(String section:sections){LinearLayout c=card(content);add(c,paragraph(section.trim()));}
        }catch(Exception e){add(content,paragraph("The full mission statement is unavailable in this installation."));}
        add(content,paragraph("Published by Genyxis Network"));
        add(content,soft("Back to Home  ›",()->show("home")));
    }
    private void privacy(){
        title(content,"Privacy and account safety");
        add(content,paragraph("Publisher: Genyxis Network. Your plan is encrypted on this phone and is not synced."));
        LinearLayout c=card(content);
        if(vaultEnvelope!=null){
            title(c,"Private Vault enabled");add(c,paragraph("Your independent passphrase protects the plan in addition to Android Keystore. The protected activity closes when it moves to the background."));
            add(c,button("Lock my private plan",()->lockVault(true)));
        }else{
            title(c,"Enable Private Vault");add(c,paragraph("Use a separate passphrase of 15 or more characters. If forgotten, your plan cannot be recovered. Keep it safely offline. Older backups may have only device-level protection."));
            EditText p=vaultPassword(c,"New passphrase"),q=vaultPassword(c,"Confirm passphrase");
            add(c,button("Enable Private Vault",()->enableVault(p,q)));
        }
        add(content,paragraph("Do not enter Social Security numbers, case passwords, or sensitive documents. Malware controlling an unlocked phone can still see what you type and read."));
        add(content,soft("Read Our Mission",()->show("mission")));
        add(content,soft("Delete all local plan data",()->new android.app.AlertDialog.Builder(this).setTitle("Delete saved plan?")
            .setMessage("This permanently erases your local plan. This is not a passphrase recovery option.")
            .setNegativeButton("Cancel",null).setPositiveButton("Delete",(d,w)->{
                if(vaultLocked||profileLocked||vaultBusy)return;
                if(!getPreferences(MODE_PRIVATE).edit().remove("private_plan").commit()){toast("Deletion failed; existing data preserved.");return;}
                savePending=false;saveHandler.removeCallbacks(deferredSave);PrivateVault.wipe(vaultKey);vaultKey=null;vaultEnvelope=null;profile=new JSONObject();
                getWindow().clearFlags(android.view.WindowManager.LayoutParams.FLAG_SECURE);show("home");
            }).show()));
    }
    private void toast(String s){Toast.makeText(this,s,Toast.LENGTH_LONG).show();}
}
