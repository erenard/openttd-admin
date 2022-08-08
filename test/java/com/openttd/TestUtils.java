package com.openttd;

import java.io.FileReader;
import java.io.IOException;
import java.util.Properties;

/**
 * Dummy class, makes samples easiers to read.
 */
public class TestUtils {

    public static void wait(int s) {
        try {
            Thread.sleep(s * 1000);
        } catch (InterruptedException ex) {
            ex.printStackTrace();
        }
    }

    public static Properties readSecrets() throws IOException {
        Properties properties = new Properties();
        properties.load(new FileReader("./openttd/secrets.cfg"));
        return properties;
    }
}
