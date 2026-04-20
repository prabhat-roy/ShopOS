package com.enterprise.admin.controller;

import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.autoconfigure.web.servlet.WebMvcTest;
import org.springframework.context.annotation.ComponentScan;
import org.springframework.context.annotation.FilterType;
import org.springframework.test.web.servlet.MockMvc;

import static org.springframework.test.web.servlet.request.MockMvcRequestBuilders.get;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.*;

@WebMvcTest(
    controllers = SystemController.class,
    excludeFilters = @ComponentScan.Filter(
        type = FilterType.ASSIGNABLE_TYPE,
        classes = com.enterprise.admin.filter.ApiKeyFilter.class
    )
)
class SystemControllerTest {

    @Autowired
    private MockMvc mockMvc;

    @Test
    void systemStats_withValidApiKey_returns200() throws Exception {
        mockMvc.perform(get("/admin/system/stats")
                .header("X-Admin-Key", "admin-secret"))
            .andExpect(status().isOk())
            .andExpect(content().contentTypeCompatibleWith("application/json"))
            .andExpect(jsonPath("$.version").exists())
            .andExpect(jsonPath("$.uptime").exists())
            .andExpect(jsonPath("$.totalTenants").exists())
            .andExpect(jsonPath("$.activeTenants").exists());
    }
}
