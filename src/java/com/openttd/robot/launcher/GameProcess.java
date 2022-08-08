package com.openttd.robot.launcher;

import java.io.BufferedReader;
import java.io.IOException;
import java.io.InputStreamReader;
import java.io.OutputStreamWriter;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

public class GameProcess extends Thread {

    private static final Logger log = LoggerFactory.getLogger(GameProcess.class);

    private final ProcessBuilder processBuilder;
    private Process process;
    private boolean running = false;

    public GameProcess(GameProcessOptions options) {
        super("GameProcess");
        log.info("Preparing new GameProcess.");
        processBuilder = new ProcessBuilder(options.getOptions());
        processBuilder.redirectErrorStream(true);
        processBuilder.directory(options.getOpenttdPath());
    }

    @Override
    public void run() {
        try {
            log.info("Starting GameProcess...");
            process = processBuilder.start();
            BufferedReader reader = new BufferedReader(new InputStreamReader(process.getInputStream()));
            while (running) {
                String line = reader.readLine();
                if (line != null) {
                    if (line.startsWith("dbg: ")) {
                        log.debug(line);
                    } else if (line.startsWith("ERROR: ")) {
                        log.error(line);
                    } else {
                        log.info(line);
                    }
                }
            }
        } catch (IOException e) {
            log.error(e.getMessage(), e);
            if (process != null) {
                log.info("Stopping GameProcess...");
                process.destroy();
                process = null;
            }
        }
    }

    public void create() throws IOException {
        running = true;
        this.start();
    }

    public void destroy() {
        running = false;
        if (process != null) {
            OutputStreamWriter osw = new OutputStreamWriter(process.getOutputStream());
            Thread t = new Thread(waitForTimeout);
            try {
                log.info("Quiting GameProcess...");
                osw.write("quit");
                osw.write("\r\n");
                osw.flush();
                t.start();
                try {
                    t.join(5000);
                    log.info("GameProcess stopped.");
                } catch (InterruptedException e) {
                    log.error(e.getMessage(), e);
                }
                process = null;
            } catch (IOException e) {
                log.error(e.getMessage(), e);
                log.info("Quiting GameProcess failed.");
            }
            if (process != null) {
                t.interrupt();
                log.info("Killing GameProcess...");
                process.destroy();
                process = null;
            }
        } else {
            log.info("GameProcess already stopped.");
        }
    }

    private Runnable waitForTimeout = new Runnable() {
        @Override
        public void run() {
            try {
                log.info("Waiting for GameProcess shutdown...");
                int exitValue = process.waitFor();
                log.info("GameProcess terminated with value: " + exitValue);
            } catch (InterruptedException ignore) {
            }
        }
    };
}
