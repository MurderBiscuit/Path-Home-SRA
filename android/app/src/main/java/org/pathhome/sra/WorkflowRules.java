package org.pathhome.sra;
import java.util.Locale;
import java.time.LocalDate;
import java.time.format.DateTimeParseException;
import java.util.regex.Pattern;
public final class WorkflowRules {
    private WorkflowRules() {}
    public static String questionTopic(String question) {
        String q=question.toLowerCase(Locale.US);
        boolean voting=q.contains("vote")||q.contains("voting")||q.contains("election");
        boolean housing=q.contains("address")||q.contains("housing")||q.contains("residence")||q.contains("reside")||q.contains("live with")||q.contains("move in")||q.contains("moving in");
        boolean reporting=q.contains("report")||q.contains("check in")||q.contains("officer")||q.contains("appointment");
        boolean travel=q.contains("travel")||q.contains("county")||q.contains("out of state")||q.contains("leave the state")||q.contains("move to another state");
        boolean employment=Pattern.compile("\\b(work|working|job|jobs|employ|employed|employment)\\b").matcher(q).find();
        int count=(voting?1:0)+(housing?1:0)+(reporting?1:0)+(travel?1:0)+(employment?1:0);
        if(count>1)return "multiple";
        if(voting)return "voting";
        if(housing)return "housing";
        if(reporting)return "reporting";
        if(travel)return "travel";
        if(employment)return "employment";
        return "";
    }
    public static boolean datesInOrder(String start,String end) {
        try {
            if(!start.isEmpty()) LocalDate.parse(start);
            if(!end.isEmpty()) LocalDate.parse(end);
            return start.isEmpty()||end.isEmpty()||start.compareTo(end)<=0;
        } catch(DateTimeParseException ex) { return false; }
    }
    public static boolean validMoney(String value){
        String s=value.trim();
        if(s.startsWith("$"))s=s.substring(1).trim();
        if(s.contains("$"))return false;
        if(s.isEmpty())return true;
        try {double v=Double.parseDouble(s);return Double.isFinite(v)&&v>=0&&v<=10000000;}catch(NumberFormatException e){return false;}
    }
    public static boolean validTime(String s){return s.matches("(?:[01]\\d|2[0-3]):[0-5]\\d");}
}
