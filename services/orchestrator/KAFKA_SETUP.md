# Kafka Setup Guide

This guide explains how to set up Kafka with Docker Compose for the orchestrator service.

## Prerequisites

- Docker and Docker Compose installed
- Ports 2181 (Zookeeper) and 9092 (Kafka) available

## Quick Start

### 1. Start Kafka with Docker Compose

From the project root directory:

```bash
docker-compose -f docker-compose.kafka.yml up -d
```

This will:
- Start Zookeeper on port 2181
- Start Kafka on port 9092
- Automatically create the required topics (email.queue, push.queue, failed.queue)

### 2. Verify Kafka is Running

```bash
# Check containers are running
docker ps | grep -E "kafka|zookeeper"

# Check Kafka health
docker exec kafka kafka-broker-api-versions --bootstrap-server localhost:9092
```

### 3. Verify Topics Were Created

```bash
# List all topics
docker exec kafka kafka-topics --list --bootstrap-server localhost:9092

# Expected output:
# __consumer_offsets
# email.queue
# push.queue
# failed.queue
```

### 4. View Topic Details

```bash
# Email queue
docker exec kafka kafka-topics --describe --bootstrap-server localhost:9092 --topic email.queue

# Push queue
docker exec kafka kafka-topics --describe --bootstrap-server localhost:9092 --topic push.queue

# Dead letter queue
docker exec kafka kafka-topics --describe --bootstrap-server localhost:9092 --topic failed.queue
```

## Topics Configuration

### email.queue
- **Purpose**: Queue for email notifications
- **Partitions**: 3
- **Retention**: 7 days
- **Consumed by**: Email Service

### push.queue
- **Purpose**: Queue for push notifications
- **Partitions**: 3
- **Retention**: 7 days
- **Consumed by**: Push Service

### failed.queue
- **Purpose**: Dead Letter Queue (DLQ) for failed messages
- **Partitions**: 3
- **Retention**: 30 days (longer retention for debugging)
- **Used by**: Email/Push services to forward failed messages

## Manual Topic Creation

If topics weren't created automatically, you can create them manually:

```bash
# Run the setup script
cd services/orchestrator
./scripts/setup-kafka-topics.sh
```

Or manually:

```bash
# Email queue
docker exec kafka kafka-topics --create \
  --bootstrap-server localhost:9092 \
  --topic email.queue \
  --partitions 3 \
  --replication-factor 1

# Push queue
docker exec kafka kafka-topics --create \
  --bootstrap-server localhost:9092 \
  --topic push.queue \
  --partitions 3 \
  --replication-factor 1

# Dead letter queue
docker exec kafka kafka-topics --create \
  --bootstrap-server localhost:9092 \
  --topic failed.queue \
  --partitions 3 \
  --replication-factor 1
```

## Testing Kafka

### 1. Produce a Test Message

```bash
# Send a message to email.queue
docker exec -it kafka kafka-console-producer.sh \
  --bootstrap-server localhost:9092 \
  --topic email.queue
```

Then type a JSON message and press Enter:
```json
{"notification_id":"test-123","notification_type":"email","user_id":"usr_123","template_code":"welcome_email","body":"Test message"}
```

Press Ctrl+C to exit.

### 2. Consume Messages

```bash
# Consume from email.queue
docker exec -it kafka kafka-console-consumer.sh \
  --bootstrap-server localhost:9092 \
  --topic email.queue \
  --from-beginning
```

### 3. Test with Orchestrator

Once your orchestrator is running:

```bash
curl -X POST http://localhost:8080/api/v1/notifications \
  -H "Content-Type: application/json" \
  -d '{
    "id": "test-123",
    "user_id": "usr_123",
    "template_code": "welcome_email",
    "notification_type": "email",
    "variables": {
      "user_name": "Test User",
      "app_name": "MyApp"
    }
  }'
```

Then check the Kafka topic:
```bash
docker exec -it kafka kafka-console-consumer.sh \
  --bootstrap-server localhost:9092 \
  --topic email.queue \
  --from-beginning \
  --max-messages 1
```

## Environment Variables

Set these environment variables for the orchestrator:

```bash
export KAFKA_BROKERS=localhost:9092
export KAFKA_EMAIL_TOPIC=email.queue
export KAFKA_PUSH_TOPIC=push.queue
export KAFKA_FAILED_TOPIC=failed.queue
```

Or create a `.env` file:
```
KAFKA_BROKERS=localhost:9092
KAFKA_EMAIL_TOPIC=email.queue
KAFKA_PUSH_TOPIC=push.queue
KAFKA_FAILED_TOPIC=failed.queue
```

## Stopping Kafka

```bash
# Stop and remove containers
docker-compose -f docker-compose.kafka.yml down

# Stop and remove containers + volumes (deletes all data)
docker-compose -f docker-compose.kafka.yml down -v
```

## Dead Letter Queue (DLQ) Usage

The `failed.queue` topic is used as a Dead Letter Queue. When the Email or Push services encounter messages they cannot process after retries, they should:

1. Log the error
2. Forward the message to `failed.queue` with error context
3. The orchestrator or monitoring system can then:
   - Alert on DLQ messages
   - Retry manually if needed
   - Analyze failure patterns

### Message Format for DLQ

When forwarding to DLQ, include the original message plus error information:

```json
{
  "original_message": { /* original KafkaNotificationPayload */ },
  "error": "Error message",
  "error_type": "processing_error",
  "failed_at": "2025-01-15T10:30:00Z",
  "retry_count": 3,
  "source_topic": "email.queue"
}
```

## Troubleshooting

### Kafka won't start
- Check if ports 2181 and 9092 are already in use
- Check Docker logs: `docker-compose -f docker-compose.kafka.yml logs`

### Topics not created
- Check kafka-init container logs: `docker logs kafka-init`
- Manually run the setup script: `./scripts/setup-kafka-topics.sh`

### Can't connect to Kafka
- Verify Kafka is running: `docker ps | grep kafka`
- Check Kafka logs: `docker logs kafka`
- Test connection: `docker exec kafka kafka-broker-api-versions --bootstrap-server localhost:9092`

### Messages not appearing
- Check producer logs in orchestrator
- Verify topic exists: `docker exec kafka kafka-topics --list --bootstrap-server localhost:9092`
- Check consumer group offsets: `docker exec kafka kafka-consumer-groups --bootstrap-server localhost:9092 --list`

## Production Considerations

For production, consider:
- Multiple Kafka brokers (replication factor > 1)
- More partitions for higher throughput
- Proper retention policies
- Monitoring and alerting
- SSL/TLS encryption
- SASL authentication
- Schema registry for message validation

