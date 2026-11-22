package org.pathhome.sra;
public class WorkflowRulesTest {
    static void check(boolean ok,String msg){if(!ok)throw new AssertionError(msg);}
    public static void main(String[] args){
        check(WorkflowRules.questionTopic("Can I move to a new address?").equals("housing"),"address -> housing");
        check(WorkflowRules.questionTopic("Can I go outside my county?").equals("travel"),"county -> travel");
        check(WorkflowRules.questionTopic("Can I report to my officer?").equals("reporting"),"officer -> reporting");
        check(WorkflowRules.questionTopic("Can I vote?").equals("voting"),"voting");
        check(WorkflowRules.questionTopic("Can I get a job?").equals("employment"),"jobs");
        check(WorkflowRules.questionTopic("Can I use a phone?").equals(""),"unknown topic must be unknown");
        check(WorkflowRules.questionTopic("Can I use a network?").equals(""),"network must not classify as employment");
        check(WorkflowRules.questionTopic("Where do I get paperwork?").equals(""),"paperwork must not classify as employment");
        check(WorkflowRules.questionTopic("Can I live with family outside my county?").equals("multiple"),"mixed residence+travel not one answer");
        check(!WorkflowRules.datesInOrder("2028-03-01","2027-03-01"),"reject backwards supervision");
        check(WorkflowRules.datesInOrder("","2028-03-01"),"unknown dates allowed");
        check(!WorkflowRules.datesInOrder("2028-02-30","2028-03-01"),"nonexistent start date rejected");
        check(!WorkflowRules.datesInOrder("2028-01-01","2028-19-09"),"invalid end date rejected");
        for(String s:new String[]{"NaN","-1","Infinity","fifty","10000001","$1$0"})check(!WorkflowRules.validMoney(s),"bad amount "+s);
        for(String s:new String[]{"","0","$21.50"})check(WorkflowRules.validMoney(s),"good amount "+s);
        check(WorkflowRules.validTime("09:00"),"09:00");
        check(!WorkflowRules.validTime("25:00"),"invalid appointment");
        System.out.println("PASS: Android pure flow rule tests");
    }
}
