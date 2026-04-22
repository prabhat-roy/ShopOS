package com.enterprise.vendoronboardingservice;

import org.springframework.web.bind.annotation.*;
import java.util.Map;

@RestController
public class HealthController {
    @GetMapping("/healthz")
    public Map<String, String> health() {
        return Map.of("status", "ok");
    }

    @GetMapping("/metrics")
    public String metrics() {
        return "# placeholder metrics\n";
    }
}
