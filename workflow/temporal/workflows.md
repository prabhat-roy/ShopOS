# Temporal Workflows — ShopOS

Temporal provides durable workflow execution for long-running, multi-step business processes.

## Namespace → Workflow Mapping

| Namespace | Workflow | Implemented In |
|---|---|---|
| shopos-commerce | `CheckoutWorkflow` | checkout-service (Go) |
| shopos-commerce | `OrderFulfillmentSaga` | saga-orchestrator (Go) |
| shopos-commerce | `PaymentRetryWorkflow` | payment-service (Java) |
| shopos-commerce | `SubscriptionBillingCycle` | subscription-billing-service (Go) |
| shopos-supply-chain | `FulfillmentWorkflow` | fulfillment-service (Go) |
| shopos-supply-chain | `ReturnLogisticsWorkflow` | returns-logistics-service (Go) |
| shopos-supply-chain | `PurchaseOrderApproval` | purchase-order-service (Kotlin) |
| shopos-financial | `PayoutWorkflow` | payout-service (Java) |
| shopos-financial | `ReconciliationWorkflow` | reconciliation-service (Kotlin) |
| shopos-financial | `KYCAMLWorkflow` | kyc-aml-service (Java) |
| shopos-communications | `NotificationDeliveryWorkflow` | notification-orchestrator (Node.js) |
| shopos-communications | `DigestSchedulerWorkflow` | digest-service (Go) |

## Key Patterns Used

### Saga / Compensating Transactions
`OrderFulfillmentSaga` orchestrates: inventory reserve → payment capture → fulfillment → shipping.
Each step has a compensating activity that rolls back on failure.

### Long-polling / Signal
`CheckoutWorkflow` waits for payment confirmation via Temporal signal from the payment provider webhook.

### Scheduled Workflows
`SubscriptionBillingCycle` runs on a cron schedule (`0 0 * * *`) per active subscription.

## Connection
Workers connect to Temporal server at `temporal-frontend.temporal-system:7233`.
Each service registers its task queue as `{service-name}-task-queue`.
