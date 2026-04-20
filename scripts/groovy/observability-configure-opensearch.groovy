def call() {
    sh """
        OS_URL=\$(grep '^OPENSEARCH_URL=' infra.env | cut -d= -f2)
        echo "Waiting for OpenSearch at \${OS_URL}..."
        until curl -sf "\${OS_URL}/_cluster/health" > /dev/null 2>&1; do sleep 10; done

        # Create ISM policy for log retention
        curl -sf -X PUT "\${OS_URL}/_plugins/_ism/policies/shopos-logs-policy" \
            -H "Content-Type: application/json" \
            -d '{
              "policy": {
                "description": "ShopOS log retention — rollover at 10gb/1d, delete after 90d",
                "default_state": "hot",
                "states": [
                  {
                    "name": "hot",
                    "actions": [{"rollover":{"min_size":"10gb","min_index_age":"1d"}}],
                    "transitions": [{"state_name":"delete","conditions":{"min_index_age":"90d"}}]
                  },
                  {
                    "name": "delete",
                    "actions": [{"delete":{}}],
                    "transitions": []
                  }
                ]
              }
            }' || true

        # Create index template for ShopOS logs
        curl -sf -X PUT "\${OS_URL}/_index_template/shopos-logs" \
            -H "Content-Type: application/json" \
            -d '{
              "index_patterns": ["shopos-logs-*"],
              "template": {
                "settings": {
                  "plugins.index_state_management.policy_id": "shopos-logs-policy",
                  "number_of_shards": 1,
                  "number_of_replicas": 0
                },
                "mappings": {
                  "properties": {
                    "@timestamp":  {"type":"date"},
                    "service":     {"type":"keyword"},
                    "level":       {"type":"keyword"},
                    "message":     {"type":"text"},
                    "trace_id":    {"type":"keyword"},
                    "span_id":     {"type":"keyword"}
                  }
                }
              }
            }' || true

        # Bootstrap alias
        curl -sf -X PUT "\${OS_URL}/shopos-logs-000001" \
            -H "Content-Type: application/json" \
            -d '{"aliases":{"shopos-logs":{"is_write_index":true}}}' || true

        echo "OpenSearch configured — ISM policy and shopos-logs template created"
    """
    echo 'opensearch configured'
}
return this
