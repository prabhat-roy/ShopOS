package com.shopos.kycamlservice;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.scheduling.annotation.EnableScheduling;

@SpringBootApplication
@EnableScheduling
public class KycAmlServiceApplication {

    public static void main(String[] args) {
        SpringApplication.run(KycAmlServiceApplication.class, args);
    }
}
