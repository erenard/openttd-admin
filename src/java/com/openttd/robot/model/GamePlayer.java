package com.openttd.robot.model;

public class GamePlayer {

    private ExternalUser externalUser;
    private double accomplishment;
    private double duration;
    private boolean winner;
    private short companyId;
    private int score;

    @Override
    public String toString() {
        StringBuilder sb = new StringBuilder();
        sb.append(externalUser.getId()).append(":").append(externalUser.getName());
        sb.append(" accomplishment:").append(accomplishment);
        sb.append(" duration:").append(duration);
        sb.append(" score:").append(score);
        sb.append(" winner:").append(winner);
        return sb.toString();
    }

    public boolean isWinner() {
        return winner;
    }

    public void setWinner(boolean winner) {
        this.winner = winner;
    }

    public ExternalUser getExternalUser() {
        return externalUser;
    }

    public void setExternalUser(ExternalUser externalUser) {
        this.externalUser = externalUser;
    }

    public double getAccomplishment() {
        return accomplishment;
    }

    public void setAccomplishment(double accomplishment) {
        this.accomplishment = accomplishment;
    }

    public double getDuration() {
        return duration;
    }

    public void setDuration(double duration) {
        this.duration = duration;
    }

    public short getCompanyId() {
        return companyId;
    }

    public void setCompanyId(short companyId) {
        this.companyId = companyId;
    }
}
