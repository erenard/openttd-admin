package com.openttd.demo;

import com.openttd.network.core.Configuration;
import java.io.FileReader;
import java.io.IOException;
import java.util.Properties;

/**
 * Dummy class, makes samples easiers to read.
 */
public class CLIUtil {
	public static void parseArguments(String [] args, Configuration configuration) {
		try {
			String url = args[0];
			configuration.host = url.split(":")[0];
			configuration.adminPort = Integer.parseInt(url.split(":")[1]);
			configuration.password = args[1];
		} catch (Exception e) {
			e.printStackTrace();
			System.err.println("Usage: java -jar openttd-robot.jar localhost:3977 admin_password");
			System.err.println("See openttd.cfg to set your server's admin_password first !");
			System.exit(0);
		}
	}

	public static void wait(int s) {
		try {
			Thread.sleep(s * 1000);
		} catch (InterruptedException ex) {
			ex.printStackTrace();
		}
	}

	public static Properties readTestProperties() throws IOException {
		Properties properties = new Properties();
		properties.load(new FileReader("./test/test.properties"));
		return properties;
	}
}
