package whatsapp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type WhatsAppService struct {
	BaseURL string
	Session string
}

func NewWhatsAppService(baseURL, session string) *WhatsAppService {
	return &WhatsAppService{BaseURL: baseURL, Session: session}
}

func (w *WhatsAppService) StartTyping(chatId string) error {
	payload := map[string]interface{}{
		"chatId":  chatId,
		"session": w.Session,
	}
	return w.post("/api/startTyping", payload)
}

func (w *WhatsAppService) SendText(chatId, text string, replyTo *string, linkPreview, linkPreviewHighQuality bool) error {
	payload := map[string]interface{}{
		"chatId":                 chatId,
		"reply_to":               replyTo,
		"text":                   text,
		"linkPreview":            linkPreview,
		"linkPreviewHighQuality": linkPreviewHighQuality,
		"session":                w.Session,
	}
	return w.post("/api/sendText", payload)
}

func (w *WhatsAppService) StopTyping(chatId string) error {
	payload := map[string]interface{}{
		"chatId":  chatId,
		"session": w.Session,
	}
	return w.post("/api/stopTyping", payload)
}

func (w *WhatsAppService) post(path string, payload map[string]interface{}) error {
	url := fmt.Sprintf("%s%s", w.BaseURL, path)
	body, _ := json.Marshal(payload)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("WhatsApp API error: %s", resp.Status)
	}
	return nil
}
