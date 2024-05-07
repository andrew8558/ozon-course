package middleware

import (
	"Homework/internal/sender"
	"io"
	"log"
	"net/http"
	"time"
)

func LogMiddleware(handler http.Handler, logSender sender.KafkaSender) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		body, err := io.ReadAll(req.Body)
		if err != nil {
			log.Println("Failed to read request body: ", err)
		}
		err = logSender.SendMessage(sender.LogMessage{
			Time:   time.Now(),
			Method: req.Method,
			Path:   req.URL.Path,
			Body:   string(body),
		})
		if err != nil {
			log.Println("Send sync message error: ", err)
		}
		handler.ServeHTTP(w, req)
	}
}
