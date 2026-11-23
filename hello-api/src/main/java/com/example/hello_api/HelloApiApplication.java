package com.example.hello_api;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.boot.builder.SpringApplicationBuilder;
import org.springframework.boot.web.servlet.support.SpringBootServletInitializer;

@SpringBootApplication(scanBasePackages = {"com.example.hello_api", "com.example.lianxi"})
public class HelloApiApplication extends SpringBootServletInitializer {

	@Override
	protected SpringApplicationBuilder configure(SpringApplicationBuilder application) {
		return application.sources(HelloApiApplication.class);
	}

	public static void main(String[] args) {
		SpringApplication.run(HelloApiApplication.class, args);
	}

}
