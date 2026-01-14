#!/bin/bash

ES_HOST="http://localhost:9200"
INDEX_NAME="rag_knowledge_chunks"
USERNAME="elastic"
PASSWORD="password" # Update with actual password if needed

echo "Creating index: $INDEX_NAME..."

curl -X PUT "$ES_HOST/$INDEX_NAME" -u "$USERNAME:$PASSWORD" -H 'Content-Type: application/json' -d'
{
  "settings": {
    "number_of_shards": 3,
    "number_of_replicas": 1
  },
  "mappings": {
    "properties": {
      "chunk_id": { "type": "keyword" },
      "doc_id": { "type": "keyword" },
      "kb_ids": { "type": "keyword" },
      "content": { 
        "type": "text", 
        "analyzer": "standard" 
      },
      "content_vector": { 
        "type": "dense_vector", 
        "dims": 768,
        "index": true,
        "similarity": "cosine"
      },
      "page_num_int": { "type": "integer" },
      "create_timestamp_flt": { "type": "float" },
      "available_int": { "type": "integer" }
    }
  }
}
'

echo -e "\nIndex created."
