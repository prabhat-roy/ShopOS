package com.shopos.invoiceservice.repository;

import com.shopos.invoiceservice.domain.Invoice;
import com.shopos.invoiceservice.domain.InvoiceStatus;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Modifying;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;

import java.time.LocalDate;
import java.util.List;
import java.util.Optional;
import java.util.UUID;

@Repository
public interface InvoiceRepository extends JpaRepository<Invoice, UUID> {

    Optional<Invoice> findByOrderId(UUID orderId);

    Page<Invoice> findByCustomerId(UUID customerId, Pageable pageable);

    List<Invoice> findByStatus(InvoiceStatus status);

    /**
     * Finds all invoices whose due date is strictly before the given date
     * and whose status is not the specified excluded status.
     * Used to detect overdue invoices (exclude already PAID/CANCELLED/VOID).
     */
    List<Invoice> findByDueDateBeforeAndStatusNot(LocalDate date, InvoiceStatus excludedStatus);

    /**
     * Finds invoices whose due date is before the given date and status is one of the provided values.
     * Used to batch-mark ISSUED and SENT invoices as OVERDUE.
     */
    @Query("SELECT i FROM Invoice i WHERE i.dueDate < :date AND i.status IN :statuses")
    List<Invoice> findOverdueEligible(@Param("date") LocalDate date,
                                      @Param("statuses") List<InvoiceStatus> statuses);

    /**
     * Bulk update: sets status = OVERDUE for all invoices that are ISSUED or SENT and past their due date.
     * Returns the count of updated rows.
     */
    @Modifying
    @Query("UPDATE Invoice i SET i.status = com.shopos.invoiceservice.domain.InvoiceStatus.OVERDUE, " +
           "i.updatedAt = CURRENT_TIMESTAMP " +
           "WHERE i.dueDate < :today " +
           "AND i.status IN (com.shopos.invoiceservice.domain.InvoiceStatus.ISSUED, " +
           "                 com.shopos.invoiceservice.domain.InvoiceStatus.SENT)")
    int bulkMarkOverdue(@Param("today") LocalDate today);

    Page<Invoice> findByCustomerIdAndStatus(UUID customerId, InvoiceStatus status, Pageable pageable);
}
