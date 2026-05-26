#!/usr/bin/env python3
"""
Генератор тестовых событий в Kafka.
Использование:
  python3 scripts/gen_events.py --count 10000 --rate 1000
"""

import argparse
import json
import random
import time
import uuid
from datetime import datetime, timezone

from kafka import KafkaProducer

QUERIES = [
    "nike sneakers", "adidas shoes", "iphone 16", "samsung galaxy",
    "macbook pro", "airpods", "jordan 1", "new balance 574",
    "levi jeans", "zara dress", "h&m sale", "gucci bag",
    "sony headphones", "ps5", "xbox series x", "nintendo switch",
    "python tutorial", "javascript course", "golang book",
    "coffee maker", "air fryer", "instant pot", "dyson vacuum",
]

SESSIONS = [f"session-{i}" for i in range(500)]
BOT_SESSIONS = ["bot-1", "bot-2"]


def make_event(query: str, session_id: str) -> dict:
    return {
        "query": query,
        "session_id": session_id,
        "timestamp": datetime.now(timezone.utc).isoformat(),
    }


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--brokers", default="localhost:9092")
    parser.add_argument("--topic", default="search-events")
    parser.add_argument("--count", type=int, default=5000, help="Total events to send")
    parser.add_argument("--rate", type=int, default=500, help="Events per second")
    parser.add_argument("--bot-ratio", type=float, default=0.1, help="Fraction of bot traffic")
    args = parser.parse_args()

    producer = KafkaProducer(
        bootstrap_servers=args.brokers,
        value_serializer=lambda v: json.dumps(v).encode("utf-8"),
    )

    interval = 1.0 / args.rate
    sent = 0

    print(f"Sending {args.count} events at {args.rate} rps to {args.brokers}/{args.topic}")
    start = time.time()

    for i in range(args.count):
        query = random.choice(QUERIES)


        if random.random() < args.bot_ratio:
            session_id = random.choice(BOT_SESSIONS)
            query = "spam query"
        else:
            session_id = random.choice(SESSIONS)

        event = make_event(query, session_id)
        producer.send(args.topic, value=event)
        sent += 1

        if sent % 1000 == 0:
            elapsed = time.time() - start
            print(f"  sent {sent}/{args.count} ({sent/elapsed:.0f} rps)")

        time.sleep(interval)

    producer.flush()
    elapsed = time.time() - start
    print(f"Done: {sent} events in {elapsed:.1f}s ({sent/elapsed:.0f} rps avg)")


if __name__ == "__main__":
    main()
