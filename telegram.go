package serviceCommunicator

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

const (
	apiURL = "https://api.telegram.org/bot"
)

type telegramStruct struct {
	Channel      int64
	BotToken     string
	Message      string
	RowData      []byte
	PhotoURL     string
	PhotoCaption string
}

func (t *telegramStruct) sendMessageToTelegram() error {
	payloads := map[string]interface{}{
		"chat_id":    t.Channel,
		"text":       t.Message,
		"parse_mode": "Markdown",
	}
	req, err := http.NewRequest("POST", apiURL+t.BotToken+"/sendMessage", nil)
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", `application/json`)
	b, marshalErr := json.Marshal(payloads)
	if marshalErr != nil {
		return marshalErr
	}
	req.Body = ioutil.NopCloser(bytes.NewBufferString(string(b)))
	resp, e := httpClient.Do(req)
	if e != nil {
		return e
	}
	_ = resp.Body.Close()
	return nil
}

func (t *telegramStruct) sendDocumentToTelegram() error {
	payloads := map[string]interface{}{
		"chat_id":  t.Channel,
		"document": t.RowData,
	}
	req, _ := http.NewRequest("POST", apiURL+t.BotToken+"/sendDocument", nil)
	req.Header.Add("Content-Type", `multipart/form-data`)
	b, marshalErr := json.Marshal(payloads)
	if marshalErr != nil {
		return marshalErr
	}
	req.Body = ioutil.NopCloser(bytes.NewBufferString(string(b)))
	resp, e := httpClient.Do(req)
	if e != nil {
		return e
	}
	_ = resp.Body.Close()
	return nil
}

func (t *telegramStruct) sendImageToTelegram() error {
	payloads := map[string]interface{}{
		"chat_id": t.Channel,
		"photo":   t.PhotoURL,
		"caption": t.PhotoCaption,
	}
	req, _ := http.NewRequest("POST", apiURL+t.BotToken+"/sendPhoto", nil)
	req.Header.Add("Content-Type", `application/json`)
	b, marshalErr := json.Marshal(payloads)
	if marshalErr != nil {
		return marshalErr
	}
	req.Body = ioutil.NopCloser(bytes.NewBufferString(string(b)))
	resp, e := httpClient.Do(req)
	if e != nil {
		return e
	}
	_ = resp.Body.Close()
	return nil
}

func sendUnavailableService(serviceName, address string) {
	text := fmt.Sprintf("Сервис '%s' по адресу %s недоступен. Обратите внимание", serviceName, address)
	fmt.Println(text)
	telegram.Channel = 219701681
	telegram.Message = text
	go func() {
		err := telegram.sendMessageToTelegram()
		if err != nil {
			logger.Error("sendUnavailableService error: %v", err)
		}
	}()
}

func sendAvailableService(serviceName, address string) {
	text := fmt.Sprintf("Сервис '%s' по адресу %s вновь доступен", serviceName, address)
	fmt.Println(text)
	telegram.Channel = 219701681
	telegram.Message = text
	go func() {
		err := telegram.sendMessageToTelegram()
		if err != nil {
			logger.Error("sendAvailableService error: %v", err)
		}
	}()
}
