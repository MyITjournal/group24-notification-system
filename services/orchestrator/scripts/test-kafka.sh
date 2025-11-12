#!/bin/bash

# Script to test Kafka connectivity and message publishing
# Usage: ./test-kafka.sh

KAFKA_CONTAINER="kafka"
BROKER="localhost:9092"
TOPIC="email.queue"

echo "Testing Kafka connection..."

# Check if Kafka container is running
if ! docker ps | grep -q "$KAFKA_CONTAINER"; then
    echo "Error: Kafka container '$KAFKA_CONTAINER' is not running"
    exit 1
fi

# Test connection
echo "1. Testing Kafka broker connection..."
if docker exec $KAFKA_CONTAINER kafka-broker-api-versions --bootstrap-server localhost:9092 > /dev/null 2>&1; then
    echo "   ✓ Kafka broker is accessible"
else
    echo "   ✗ Cannot connect to Kafka broker"
    exit 1
fi

# List topics
echo ""
echo "2. Listing topics..."
docker exec $KAFKA_CONTAINER kafka-topics --list --bootstrap-server localhost:9092

# Check if required topics exist
echo ""
echo "3. Checking required topics..."
for topic in "email.queue" "push.queue" "failed.queue"; do
    if docker exec $KAFKA_CONTAINER kafka-topics --list --bootstrap-server localhost:9092 | grep -q "^${topic}$"; then
        echo "   ✓ Topic '$topic' exists"
    else
        echo "   ✗ Topic '$topic' does not exist"
    fi
done

# Test message production
echo ""
echo "4. Testing message production..."
TEST_MESSAGE='{"notification_id":"test-123","notification_type":"email","user_id":"usr_test","template_code":"test","body":"Test message"}'

echo "$TEST_MESSAGE" | docker exec -i $KAFKA_CONTAINER kafka-console-producer.sh \
    --bootstrap-server localhost:9092 \
    --topic $TOPIC > /dev/null 2>&1

if [ $? -eq 0 ]; then
    echo "   ✓ Message published successfully"
else
    echo "   ✗ Failed to publish message"
    exit 1
fi

# Test message consumption
echo ""
echo "5. Testing message consumption..."
CONSUMED=$(timeout 5 docker exec $KAFKA_CONTAINER kafka-console-consumer.sh \
    --bootstrap-server localhost:9092 \
    --topic $TOPIC \
    --from-beginning \
    --max-messages 1 2>/dev/null)

if [ -n "$CONSUMED" ]; then
    echo "   ✓ Message consumed successfully"
    echo "   Message: $CONSUMED"
else
    echo "   ✗ Failed to consume message"
    exit 1
fi

echo ""
echo "All Kafka tests passed! ✓"

