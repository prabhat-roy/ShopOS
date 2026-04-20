def call() {
    sh """
        ES_URL=\$(grep '^ELASTICSEARCH_URL=' infra.env | cut -d= -f2)
        echo "Waiting for Elasticsearch at \${ES_URL}..."
        until curl -sf "\${ES_URL}/_cluster/health" > /dev/null 2>&1; do sleep 10; done

        # Create ILM policy for log retention (30-day hot, delete after 90 days)
        curl -sf -X PUT "\${ES_URL}/_ilm/policy/shopos-logs-policy" \
            -H "Content-Type: application/json" \
            -d '{
              "policy": {
                "phases": {
                  "hot":    {"min_age":"0ms","actions":{"rollover":{"max_size":"10gb","max_age":"1d"}}},
                  "warm":   {"min_age":"7d","actions":{"shrink":{"number_of_shards":1},"forcemerge":{"max_num_segments":1}}},
                  "delete": {"min_age":"90d","actions":{"delete":{}}}
                }
              }
            }' || true

        # Create index template for ShopOS service logs
        curl -sf -X PUT "\${ES_URL}/_index_template/shopos-logs" \
            -H "Content-Type: application/json" \
            -d '{
              "index_patterns": ["shopos-logs-*"],
              "template": {
                "settings": {
                  "index.lifecycle.name": "shopos-logs-policy",
                  "index.lifecycle.rollover_alias": "shopos-logs",
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

        # Bootstrap the alias with a write index
        curl -sf -X PUT "\${ES_URL}/shopos-logs-000001" \
            -H "Content-Type: application/json" \
            -d '{"aliases":{"shopos-logs":{"is_write_index":true}}}' || true

        echo "Elasticsearch configured — ILM policy and shopos-logs template created"
    """
    echo 'elasticsearch configured'
}
return this
