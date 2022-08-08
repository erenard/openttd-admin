package com.openttd.server;

import java.io.BufferedReader;
import java.io.IOException;
import java.io.InputStreamReader;
import java.net.ServerSocket;
import org.junit.jupiter.api.AfterAll;
import org.junit.jupiter.api.BeforeAll;

/**
 *
 * @author eric
 */
public class OpenttdServerWrapper {

    private static OpenttdDedicatedServer openttdServer;

    private static boolean checkAdminPort() {
        try ( ServerSocket socket = new ServerSocket(3977)) {
            // Test the admin port
            socket.close();
            return false;
        } catch (IOException ioe) {
            System.err.println(ioe.getMessage());
            System.err.println("openttd -D already running");
            return true;
        }
    }

    @BeforeAll
    public static void startServer() throws IOException {
        if (!checkAdminPort()) {
            openttdServer = new OpenttdDedicatedServer();
        }
    }

    @AfterAll
    public static void stopServer() {
        if (openttdServer != null) {
            openttdServer.destroy();
        }
    }

    static class OpenttdDedicatedServer {

        private final Process server;
        private boolean running = true;
        private final Thread consoleStream;
        private final InputStreamReader isr;
        private final BufferedReader br;

        private OpenttdDedicatedServer() throws IOException {
            // openttd.cfg will also silently load secrets.cfg
            ProcessBuilder processBuilder = new ProcessBuilder("openttd", "-D", "-d", "9,console=1,sprite=0,grf=0,sl=0", "-c", "openttd/openttd.cfg");
            processBuilder.redirectErrorStream(true);
            this.server = processBuilder.start();
            isr = new InputStreamReader(this.server.getInputStream());
            br = new BufferedReader(isr);
            this.consoleStream = new Thread(() -> {
                while (running) {
                    try {
                        String line = br.readLine();
                        if (line != null) {
                            System.out.println(line);
                        }
                    } catch (IOException ioe) {
                        System.err.println(ioe.getMessage());
                        running = false;
                    }
                }
            });
            this.consoleStream.start();
        }

        public void destroy() {
            running = false;
            try {
                this.consoleStream.join();
            } catch (InterruptedException ie) {
                this.consoleStream.interrupt();
            }
            try {
                br.close();
            } catch (IOException ioe) {
                System.err.println(ioe.getMessage());
            }
            try {
                isr.close();
            } catch (IOException ioe) {
                System.err.println(ioe.getMessage());
            }
            this.server.destroy();
            this.server.destroyForcibly();
        }
    }
}
