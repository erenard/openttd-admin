package com.openttd;

import java.io.FileReader;
import java.io.IOException;
import java.util.Collection;
import java.util.Properties;
import java.util.function.Supplier;

/**
 * Dummy class, makes samples easiers to read.
 */
public class TestUtils {

    /** Poll interval used by await(). */
    private static final int POLL_MILLIS = 50;

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

    /**
     * Wait for an asynchronous event instead of sleeping a fixed amount of
     * time: returns as soon as the condition holds.
     *
     * @param condition checked every POLL_MILLIS
     * @param timeoutSeconds
     * @param description what is awaited, used in the failure message
     * @throws AssertionError if the condition never held
     */
    public static void await(Supplier<Boolean> condition, int timeoutSeconds, String description) {
        long deadline = System.currentTimeMillis() + timeoutSeconds * 1000L;
        while (System.currentTimeMillis() < deadline) {
            if (Boolean.TRUE.equals(condition.get())) {
                return;
            }
            try {
                Thread.sleep(POLL_MILLIS);
            } catch (InterruptedException e) {
                throw new AssertionError("Interrupted while waiting for " + description, e);
            }
        }
        throw new AssertionError("Timed out after " + timeoutSeconds + "s waiting for " + description);
    }

    /**
     * Wait for a collection to hold exactly `size` events.
     */
    public static void awaitSize(Collection<?> collection, int size, int timeoutSeconds, String description) {
        await(() -> collection.size() >= size, timeoutSeconds,
                description + " (" + size + " events, was " + collection.size() + ")");
    }
}
