package com.enterprise.auctionsettlement;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RestController;

import java.util.Map;

@SpringBootApplication
public class Application {

    public static void main(String[] args) {
        SpringApplication.run(Application.class, args);
    }

    @RestController
    public static class HealthController {
        @GetMapping("/healthz")
        public Map<String, String> healthz() {
            return Map.of("status", "ok");
        }
    }
}
