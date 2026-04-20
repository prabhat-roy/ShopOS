package com.shopos.erpintegrationservice.dto;

import com.shopos.erpintegrationservice.domain.SyncStatus;

import java.time.Instant;
import java.util.List;
import java.util.UUID;

/**
 * Response returned after initiating or completing a synchronisation operation.
 *
 * @param syncId           Unique identifier for this sync operation, used for status polling.
 * @param status           Current state of the sync operation.
 * @param recordsProcessed Number of records (orders or inventory lines) successfully processed.
 * @param errors           List of human-readable error messages for any records that failed.
 * @param completedAt      Timestamp when processing finished; null if still in progress.
 */
public record SyncResponse(
        UUID syncId,
        SyncStatus status,
        int recordsProcessed,
        List<String> errors,
        Instant completedAt
) {}
