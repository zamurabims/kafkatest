package consumer

import (
	"context"
	"encoding/json"
	"log"
	"strings"
	"time"

	kafka "github.com/segmentio/kafka-go"

	"wish1/internal/metrics"
	"wish1/internal/stoplist"
	"wish1/internal/window"
)

type Event struct {
	Query     string    `json:"query"`
	SessionID string    `json:"session_id"`
	Timestamp time.Time `json:"timestamp"`
}

type Consumer struct {
	reader   *kafka.Reader
	win      *window.Window
	stoplist *stoplist.StopList
}

func New(brokers, topic, _ string, win *window.Window, sl *stoplist.StopList) *Consumer {
	log.Printf("consumer: connecting to brokers=%s topic=%s", brokers, topic)

	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     strings.Split(brokers, ","),
		Topic:       topic,
		Partition:   0,
		MinBytes:    1,
		MaxBytes:    10e6,
		MaxWait:     3 * time.Second,
		StartOffset: kafka.LastOffset,
		ErrorLogger: kafka.LoggerFunc(func(msg string, args ...interface{}) {
			log.Printf("[kafka ERROR] "+msg, args...)
		}),
	})
	return &Consumer{reader: r, win: win, stoplist: sl}
}

func (c *Consumer) Run(ctx context.Context) {
	log.Printf("consumer: started, waiting for messages...")
	for {
		msg, err := c.reader.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			log.Printf("consumer read error: %v", err)
			time.Sleep(time.Second)
			continue
		}

		log.Printf("consumer: got message: %s", string(msg.Value))

		var ev Event
		if err := json.Unmarshal(msg.Value, &ev); err != nil {
			log.Printf("consumer unmarshal error: %v", err)
			metrics.RecordsDropped.WithLabelValues("invalid").Inc()
			continue
		}

		query := normalize(ev.Query)
		if query == "" || ev.SessionID == "" {
			metrics.RecordsDropped.WithLabelValues("invalid").Inc()
			continue
		}
		if c.stoplist.Contains(query) {
			metrics.RecordsDropped.WithLabelValues("stoplist").Inc()
			continue
		}

		metrics.RecordsTotal.Inc()

		ts := ev.Timestamp
		if ts.IsZero() {
			ts = time.Now()
		}

		c.win.Record(query, ev.SessionID, ts)
		log.Printf("consumer: recorded query=%s session=%s", query, ev.SessionID)
	}
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}

func normalize(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}
