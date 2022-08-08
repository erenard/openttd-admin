package com.openttd.robot.launcher;

import java.io.File;
import java.util.ArrayList;
import java.util.List;

public class GameProcessOptions {

    private final File openttdPath;
    private final List<String> options;

    public GameProcessOptions(String openttdPath, String openttdCommand) {
        this.openttdPath = new File(openttdPath);
        options = new ArrayList<>();
        options.add(openttdCommand);
        options.add("-D");
        options.add("-d 1");
    }

    public void addOption(String option) {
        options.add(option);
    }

    public File getOpenttdPath() {
        return openttdPath;
    }

    public List<String> getOptions() {
        return options;
    }
}
