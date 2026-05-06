package webhook

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
)

type WebHookService struct {
	gateways []string
}

func NewWebHookService(gateways []string) *WebHookService {
	return &WebHookService{gateways: gateways}
}

func (w *WebHookService) Send(payload interface{}) {
	data, err := json.Marshal(payload)
	if err != nil {
		log.Println("webhook marshal error:", err)
		return
	}
	for _, url := range w.gateways {
		go func(url string) {
			resp, err := http.Post(url, "application/json", bytes.NewBuffer(data))
			if err != nil {
				log.Printf("webhook POST error to %s: %v", url, err)
				return
			}
			defer resp.Body.Close()
			if resp.StatusCode >= 300 {
				log.Printf("webhook POST non-2xx to %s: %d", url, resp.StatusCode)
			}
		}(url)
	}
}
