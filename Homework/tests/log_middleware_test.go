//go:build integration
// +build integration

package tests

import (
	"Homework/internal/infrastructure/kafka"
	"Homework/internal/middleware"
	"Homework/internal/receiver"
	"Homework/internal/sender"
	"bytes"
	"encoding/json"
	"github.com/IBM/sarama"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func Test_LogMiddleware(t *testing.T) {
	var brokers = []string{"127.0.0.1:9081"}
	t.Run("succes log's work", func(t *testing.T) {
		//arrange
		kafkaProducer, err := kafka.NewProducer(brokers)
		require.NoError(t, err)

		kafkaSender := sender.NewKafkaSender(kafkaProducer, "test_logs")

		kafkaConsumer, err := kafka.NewConsumer(brokers)
		require.NoError(t, err)

		ch := make(chan sender.LogMessage)
		handlers := map[string]receiver.HandleFunc{
			"test_logs": func(message *sarama.ConsumerMessage) {
				var pm sender.LogMessage
				err = json.Unmarshal(message.Value, &pm)
				require.NoError(t, err)
				ch <- pm
			},
		}

		logReceiver := receiver.NewReceiver(kafkaConsumer, handlers)
		logReceiver.Subscribe("test_logs")

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
		handlerWithLogging := middleware.LogMiddleware(handler, *kafkaSender)

		//act
		reqBody := []byte("test request body")
		req, err := http.NewRequest("POST", "/test", bytes.NewBuffer(reqBody))
		require.NoError(t, err)
		w := httptest.NewRecorder()
		reqTime := time.Now()
		handlerWithLogging.ServeHTTP(w, req)

		//assert
		gotLog := <-ch
		assert.Equal(t, reqTime.Truncate(time.Second), gotLog.Time.Truncate(time.Second))
		assert.Equal(t, "POST", gotLog.Method)
		assert.Equal(t, "/test", gotLog.Path)
		assert.Equal(t, "test request body", gotLog.Body)
	})
}
