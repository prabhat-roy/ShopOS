package com.enterprise.lease

import org.springframework.boot.autoconfigure.SpringBootApplication
import org.springframework.boot.runApplication
import org.springframework.web.bind.annotation.GetMapping
import org.springframework.web.bind.annotation.RestController

@SpringBootApplication
class Application

fun main(args: Array<String>) {
    runApplication<Application>(*args)
}

@RestController
class HealthController {
    @GetMapping("/healthz")
    fun healthz(): Map<String, String> = mapOf("status" to "ok")
}
