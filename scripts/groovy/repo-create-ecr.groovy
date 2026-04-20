def call() {
    def envVars = readFile('infra.env').trim().split('\n').collectEntries { line ->
        def parts = line.split('=', 2)
        parts.length == 2 ? [(parts[0]): parts[1]] : [:]
    }

    def region = sh(script: """
        AWS_REGION=\$(aws configure get region 2>/dev/null || echo '')
        if [ -z "\$AWS_REGION" ]; then
            TOKEN=\$(curl -s -X PUT "http://169.254.169.254/latest/api/token" -H "X-aws-ec2-metadata-token-ttl-seconds: 21600")
            AWS_REGION=\$(curl -s -H "X-aws-ec2-metadata-token: \$TOKEN" http://169.254.169.254/latest/meta-data/placement/region)
        fi
        echo "\$AWS_REGION"
    """, returnStdout: true).trim()

    def accountId = sh(script: "aws sts get-caller-identity --query Account --output text", returnStdout: true).trim()

    def repos = [
        'api-gateway', 'web-bff', 'mobile-bff', 'partner-bff', 'config-service',
        'feature-flag-service', 'rate-limiter-service', 'health-check-service',
        'saga-orchestrator', 'event-store-service', 'cache-warming-service',
        'webhook-service', 'scheduler-service', 'worker-job-queue', 'audit-service',
        'load-generator', 'admin-portal', 'graphql-gateway', 'dead-letter-service',
        'geolocation-service', 'event-replay-service', 'tenant-service',
        'auth-service', 'user-service', 'session-service', 'permission-service',
        'mfa-service', 'gdpr-service', 'api-key-service', 'device-fingerprint-service',
        'product-catalog-service', 'category-service', 'brand-service', 'pricing-service',
        'inventory-service', 'bundle-service', 'configurator-service',
        'subscription-product-service', 'search-service', 'seo-service',
        'product-import-service', 'price-list-service',
        'cart-service', 'checkout-service', 'order-service', 'payment-service',
        'shipping-service', 'currency-service', 'tax-service', 'promotions-service',
        'loyalty-service', 'return-refund-service', 'subscription-billing-service',
        'fraud-detection-service', 'wallet-service', 'ab-testing-service',
        'gift-card-service', 'address-validation-service', 'digital-goods-service',
        'voucher-service', 'pre-order-service', 'backorder-service', 'waitlist-service',
        'flash-sale-service', 'bnpl-service',
        'vendor-service', 'purchase-order-service', 'warehouse-service',
        'fulfillment-service', 'tracking-service', 'label-service',
        'carrier-integration-service', 'demand-forecast-service', 'customs-duties-service',
        'returns-logistics-service', 'supplier-portal-service', 'cold-chain-service',
        'supplier-rating-service',
        'invoice-service', 'accounting-service', 'payout-service',
        'reconciliation-service', 'tax-reporting-service', 'expense-management-service',
        'credit-service', 'kyc-aml-service', 'budget-service', 'chargeback-service',
        'revenue-recognition-service',
        'review-rating-service', 'qa-service', 'wishlist-service', 'compare-service',
        'recently-viewed-service', 'support-ticket-service', 'live-chat-service',
        'consent-management-service', 'age-verification-service', 'survey-service',
        'feedback-service', 'price-alert-service', 'back-in-stock-service',
        'gift-registry-service',
        'notification-orchestrator', 'email-service', 'sms-service',
        'push-notification-service', 'template-service', 'in-app-notification-service',
        'digest-service', 'whatsapp-service', 'chatbot-service',
        'media-asset-service', 'image-processing-service', 'document-service',
        'cms-service', 'video-service', 'sitemap-service', 'i18n-l10n-service',
        'data-export-service',
        'analytics-service', 'reporting-service', 'recommendation-service',
        'sentiment-analysis-service', 'price-optimization-service', 'ml-feature-store',
        'personalization-service', 'data-pipeline-service', 'ad-service',
        'event-tracking-service', 'attribution-service', 'clv-service',
        'search-analytics-service',
        'organization-service', 'contract-service', 'quote-rfq-service',
        'approval-workflow-service', 'b2b-credit-limit-service', 'edi-service',
        'marketplace-seller-service',
        'erp-integration-service', 'marketplace-connector-service',
        'social-commerce-service', 'crm-integration-service',
        'payment-gateway-integration', 'logistics-provider-integration',
        'tax-provider-integration', 'pim-integration-service', 'cdp-integration-service',
        'accounting-integration-service',
        'affiliate-service', 'referral-service', 'influencer-service',
        'commission-payout-service'
    ]

    repos.each { repo ->
        def existing = sh(
            script: "aws ecr describe-repositories --repository-names shopos/${repo} --region ${region} 2>&1 || true",
            returnStdout: true
        ).trim()
        if (!existing.contains('repositoryUri')) {
            sh "aws ecr create-repository --repository-name shopos/${repo} --region ${region} --image-scanning-configuration scanOnPush=true"
            echo "Created ECR repo: shopos/${repo}"
        } else {
            echo "ECR repo already exists: shopos/${repo}"
        }
    }

    def registryUrl = "${accountId}.dkr.ecr.${region}.amazonaws.com"
    sh "echo 'ECR_REGISTRY=${registryUrl}' >> infra.env"
    sh "echo 'ECR_REGION=${region}' >> infra.env"
    echo "ECR registry: ${registryUrl}"
}

return this
