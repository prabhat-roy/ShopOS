package com.enterprise.purchaserequisitionservice

import org.springframework.web.bind.annotation.*

@RestController
class HealthController {
    @GetMapping("/healthz")
    fun health() = mapOf("status" to "ok")

    @GetMapping("/metrics")
    fun metrics() = "# placeholder metrics\n"
}
