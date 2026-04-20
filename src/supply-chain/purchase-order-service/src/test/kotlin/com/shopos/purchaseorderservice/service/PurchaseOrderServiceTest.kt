package com.shopos.purchaseorderservice.service

import com.shopos.purchaseorderservice.domain.POStatus
import com.shopos.purchaseorderservice.domain.PurchaseOrder
import com.shopos.purchaseorderservice.domain.PurchaseOrderItem
import com.shopos.purchaseorderservice.dto.CreatePORequest
import com.shopos.purchaseorderservice.dto.ItemReceiptEntry
import com.shopos.purchaseorderservice.dto.POItemRequest
import com.shopos.purchaseorderservice.dto.ReceiveItemsRequest
import com.shopos.purchaseorderservice.exception.InvalidPOTransitionException
import com.shopos.purchaseorderservice.exception.PurchaseOrderNotFoundException
import com.shopos.purchaseorderservice.repository.PurchaseOrderRepository
import org.assertj.core.api.Assertions.assertThat
import org.assertj.core.api.Assertions.assertThatThrownBy
import org.junit.jupiter.api.BeforeEach
import org.junit.jupiter.api.DisplayName
import org.junit.jupiter.api.Test
import org.junit.jupiter.api.extension.ExtendWith
import org.mockito.InjectMocks
import org.mockito.Mock
import org.mockito.junit.jupiter.MockitoExtension
import org.mockito.kotlin.any
import org.mockito.kotlin.never
import org.mockito.kotlin.verify
import org.mockito.kotlin.whenever
import java.math.BigDecimal
import java.time.Instant
import java.time.LocalDate
import java.util.Optional
import java.util.UUID

@ExtendWith(MockitoExtension::class)
class PurchaseOrderServiceTest {

    @Mock
    private lateinit var purchaseOrderRepository: PurchaseOrderRepository

    @InjectMocks
    private lateinit var purchaseOrderService: PurchaseOrderService

    private lateinit var poId: UUID
    private lateinit var vendorId: UUID
    private lateinit var itemId: UUID
    private lateinit var sampleItem: PurchaseOrderItem
    private lateinit var samplePO: PurchaseOrder

    @BeforeEach
    fun setUp() {
        poId = UUID.randomUUID()
        vendorId = UUID.randomUUID()
        itemId = UUID.randomUUID()

        sampleItem = PurchaseOrderItem(
            id = itemId,
            orderId = poId,
            productId = "PROD-001",
            sku = "SKU-ABC-001",
            productName = "Industrial Widget",
            quantity = 100,
            unitPrice = BigDecimal("9.99"),
            totalPrice = BigDecimal("999.00"),
            receivedQty = 0
        )

        samplePO = PurchaseOrder(
            id = poId,
            vendorId = vendorId,
            status = POStatus.DRAFT,
            totalAmount = BigDecimal("999.00"),
            currency = "USD",
            notes = "Urgent order",
            items = mutableListOf(sampleItem),
            expectedDelivery = LocalDate.now().plusDays(14),
            createdAt = Instant.now(),
            updatedAt = Instant.now()
        )
    }

    @Test
    @DisplayName("createPO - success: computes totals and persists with DRAFT status")
    fun createPO_success() {
        val request = CreatePORequest(
            vendorId = vendorId,
            items = listOf(
                POItemRequest(
                    productId = "PROD-001",
                    sku = "SKU-ABC-001",
                    productName = "Industrial Widget",
                    quantity = 100,
                    unitPrice = BigDecimal("9.99")
                )
            ),
            notes = "Urgent order",
            expectedDelivery = LocalDate.now().plusDays(14)
        )

        whenever(purchaseOrderRepository.save(any())).thenAnswer { invocation ->
            val po = invocation.getArgument<PurchaseOrder>(0)
            po.id = poId
            po.createdAt = Instant.now()
            po.updatedAt = Instant.now()
            po
        }

        val response = purchaseOrderService.createPO(request)

        assertThat(response).isNotNull
        assertThat(response.vendorId).isEqualTo(vendorId)
        assertThat(response.status).isEqualTo(POStatus.DRAFT)
        assertThat(response.totalAmount).isEqualByComparingTo(BigDecimal("999.00"))
        assertThat(response.items).hasSize(1)
        verify(purchaseOrderRepository, org.mockito.kotlin.atLeast(1)).save(any())
    }

    @Test
    @DisplayName("getPO - found: returns correct POResponse")
    fun getPO_found() {
        whenever(purchaseOrderRepository.findById(poId)).thenReturn(Optional.of(samplePO))

        val response = purchaseOrderService.getPO(poId)

        assertThat(response.id).isEqualTo(poId)
        assertThat(response.vendorId).isEqualTo(vendorId)
        assertThat(response.status).isEqualTo(POStatus.DRAFT)
        verify(purchaseOrderRepository).findById(poId)
    }

    @Test
    @DisplayName("getPO - not found: throws PurchaseOrderNotFoundException")
    fun getPO_notFound() {
        val unknownId = UUID.randomUUID()
        whenever(purchaseOrderRepository.findById(unknownId)).thenReturn(Optional.empty())

        assertThatThrownBy { purchaseOrderService.getPO(unknownId) }
            .isInstanceOf(PurchaseOrderNotFoundException::class.java)
            .hasMessageContaining(unknownId.toString())
    }

    @Test
    @DisplayName("listPOs - no filters: returns all purchase orders")
    fun listPOs_noFilters_returnsAll() {
        val second = PurchaseOrder(
            id = UUID.randomUUID(),
            vendorId = UUID.randomUUID(),
            status = POStatus.APPROVED,
            totalAmount = BigDecimal("250.00"),
            currency = "USD",
            createdAt = Instant.now(),
            updatedAt = Instant.now()
        )
        whenever(purchaseOrderRepository.findAll()).thenReturn(listOf(samplePO, second))

        val results = purchaseOrderService.listPOs(null, null)

        assertThat(results).hasSize(2)
        verify(purchaseOrderRepository).findAll()
        verify(purchaseOrderRepository, never()).findByVendorId(any())
        verify(purchaseOrderRepository, never()).findByStatus(any())
    }

    @Test
    @DisplayName("submitPO - DRAFT -> SUBMITTED: succeeds")
    fun submitPO_fromDraft_succeeds() {
        whenever(purchaseOrderRepository.findById(poId)).thenReturn(Optional.of(samplePO))
        whenever(purchaseOrderRepository.save(any())).thenReturn(samplePO)

        purchaseOrderService.submitPO(poId)

        assertThat(samplePO.status).isEqualTo(POStatus.SUBMITTED)
        verify(purchaseOrderRepository).save(samplePO)
    }

    @Test
    @DisplayName("submitPO - already SUBMITTED: throws InvalidPOTransitionException")
    fun submitPO_alreadySubmitted_throwsException() {
        samplePO.status = POStatus.SUBMITTED
        whenever(purchaseOrderRepository.findById(poId)).thenReturn(Optional.of(samplePO))

        assertThatThrownBy { purchaseOrderService.submitPO(poId) }
            .isInstanceOf(InvalidPOTransitionException::class.java)
            .hasMessageContaining("SUBMITTED")
    }

    @Test
    @DisplayName("approvePO - SUBMITTED -> APPROVED: succeeds")
    fun approvePO_fromSubmitted_succeeds() {
        samplePO.status = POStatus.SUBMITTED
        whenever(purchaseOrderRepository.findById(poId)).thenReturn(Optional.of(samplePO))
        whenever(purchaseOrderRepository.save(any())).thenReturn(samplePO)

        purchaseOrderService.approvePO(poId)

        assertThat(samplePO.status).isEqualTo(POStatus.APPROVED)
        verify(purchaseOrderRepository).save(samplePO)
    }

    @Test
    @DisplayName("receiveItems - partial receipt: sets status to PARTIALLY_RECEIVED")
    fun receiveItems_partial_setsPartiallyReceived() {
        samplePO.status = POStatus.APPROVED
        whenever(purchaseOrderRepository.findById(poId)).thenReturn(Optional.of(samplePO))
        whenever(purchaseOrderRepository.save(any())).thenReturn(samplePO)

        val request = ReceiveItemsRequest(
            receipts = listOf(ItemReceiptEntry(itemId = itemId, receivedQty = 50))
        )

        val response = purchaseOrderService.receiveItems(poId, request)

        assertThat(samplePO.status).isEqualTo(POStatus.PARTIALLY_RECEIVED)
        assertThat(sampleItem.receivedQty).isEqualTo(50)
        verify(purchaseOrderRepository).save(samplePO)
    }

    @Test
    @DisplayName("receiveItems - full receipt: sets status to FULLY_RECEIVED")
    fun receiveItems_full_setsFullyReceived() {
        samplePO.status = POStatus.APPROVED
        whenever(purchaseOrderRepository.findById(poId)).thenReturn(Optional.of(samplePO))
        whenever(purchaseOrderRepository.save(any())).thenReturn(samplePO)

        val request = ReceiveItemsRequest(
            receipts = listOf(ItemReceiptEntry(itemId = itemId, receivedQty = 100))
        )

        purchaseOrderService.receiveItems(poId, request)

        assertThat(samplePO.status).isEqualTo(POStatus.FULLY_RECEIVED)
        assertThat(sampleItem.receivedQty).isEqualTo(100)
    }
}
