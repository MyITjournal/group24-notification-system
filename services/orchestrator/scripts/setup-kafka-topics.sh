#!/bin/bash

# Script to manually create Kafka topics if needed
# Usage: ./setup-kafka-topics.sh

KAFKA_CONTAINER="kafka"
BROKER="localhost:9092"

echo "Setting up Kafka topics..."

# Check if Kafka container is running
if ! docker ps | grep -q "$KAFKA_CONTAINER"; then
    echo "Error: Kafka container '$KAFKA_CONTAINER' is not running"
    echo "Please start Kafka first with: docker-compose -f docker-compose.kafka.yml up -d"
    exit 1
fi

# Wait for Kafka to be ready
echo "Waiting for Kafka to be ready..."
timeout=30
counter=0
until docker exec $KAFKA_CONTAINER kafka-broker-api-versions --bootstrap-server localhost:9092 > /dev/null 2>&1; do
    if [ $counter -ge $timeout ]; then
        echo "Error: Kafka did not become ready in time"
        exit 1
    fi
    echo "Waiting for Kafka... ($counter/$timeout)"
    sleep 2
    counter=$((counter + 2))
done

echo "Kafka is ready!"

# Create email.queue topic
echo "Creating email.queue topic..."
docker exec $KAFKA_CONTAINER kafka-topics --create --if-not-exists \
    --bootstrap-server localhost:9092 \
    --topic email.queue \
    --partitions 3 \
    --replication-factor 1 \
    --config retention.ms=604800000 \
    --config segment.ms=86400000

# Create push.queue topic
echo "Creating push.queue topic..."
docker exec $KAFKA_CONTAINER kafka-topics --create --if-not-exists \
    --bootstrap-server localhost:9092 \
    --topic push.queue \
    --partitions 3 \
    --replication-factor 1 \
    --config retention.ms=604800000 \
    --config segment.ms=86400000

# Create failed.queue topic (Dead Letter Queue)
echo "Creating failed.queue topic (Dead Letter Queue)..."
docker exec $KAFKA_CONTAINER kafka-topics --create --if-not-exists \
    --bootstrap-server localhost:9092 \
    --topic failed.queue \
    --partitions 3 \
    --replication-factor 1 \
    --config retention.ms=2592000000 \
    --config segment.ms=86400000

# List all topics
echo ""
echo "Created topics:"
docker exec $KAFKA_CONTAINER kafka-topics --list --bootstrap-server localhost:9092

echo ""
echo "Topic details:"
docker exec $KAFKA_CONTAINER kafka-topics --describe --bootstrap-server localhost:9092 --topic email.queue
docker exec $KAFKA_CONTAINER kafka-topics --describe --bootstrap-server localhost:9092 --topic push.queue
docker exec $KAFKA_CONTAINER kafka-topics --describe --bootstrap-server localhost:9092 --topic failed.queue

echo ""
echo "Kafka topics setup complete!"

