package webex

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"go.uber.org/zap"
)

type WebexConfiguration struct {
	Log      *zap.Logger
	ApiUrl   string
	ApiToken string
	SpaceId  string
	Message  string
	Card     string
}

type MessageCreateRequest struct {
	RoomID        string                `json:"roomId,omitempty"`        // Room ID.
	ParentID      string                `json:"parentId,omitempty"`      // Parent ID
	ToPersonID    string                `json:"toPersonId,omitempty"`    // Person ID (for type=direct).
	ToPersonEmail string                `json:"toPersonEmail,omitempty"` // Person email (for type=direct).
	Text          string                `json:"text,omitempty"`          // Message in plain text format.
	Markdown      string                `json:"markdown,omitempty"`      // Message in markdown format.
	Attachments   []WebexCardAttachment `json:"attachments,omitempty"`
}

type WebexCardAttachment struct {
	ContentType string                 `json:"contentType,omitempty"`
	Content     map[string]interface{} `json:"content,omitempty"`
}

func checkWebexHttpResponse(resp *http.Response, err error) error {
	respBodyString := ""
	if resp != nil && resp.Body != nil {
		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)
		respBodyString = buf.String()
	}
	if err != nil {

		respBodyString = fmt.Sprintf("%s - %v", respBodyString, err)

		return fmt.Errorf("unable to send webex message: Response Body: %v", respBodyString)
	}

	if resp.StatusCode <= 199 || resp.StatusCode >= 400 {
		return fmt.Errorf("webex teams API call failed. Status Code: %v. Response Body: %v", resp.StatusCode, respBodyString)
	}

	return nil
}

func (c WebexConfiguration) SendWithCard() error {

	log := c.Log.With(zap.String("apiUrl", c.ApiUrl), zap.String("spaceId", c.SpaceId)).Sugar()

	cardObject := map[string]interface{}{}
	err := json.Unmarshal([]byte(c.Card), &cardObject)
	if err != nil {
		return err
	}
	var cardAttach []WebexCardAttachment
	cardData := WebexCardAttachment{
		ContentType: "application/vnd.microsoft.card.adaptive",
		Content:     cardObject,
	}
	cardAttach = append(cardAttach, cardData)
	body := MessageCreateRequest{
		RoomID:      c.SpaceId,
		Text:        c.Message,
		Attachments: cardAttach,
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal the webex request body into json - %v", err)
	}

	return c.MakeApiCall(bodyBytes, log)

}

func (c WebexConfiguration) SendMessage() error {
	log := c.Log.With(zap.String("apiUrl", c.ApiUrl), zap.String("spaceId", c.SpaceId)).Sugar()

	body := MessageCreateRequest{
		RoomID: c.SpaceId,
		Text:   c.Message,
	}

	bodyBytes, _ := json.Marshal(body)

	return c.MakeApiCall(bodyBytes, log)

}

func (c WebexConfiguration) MakeApiCall(bodyBytes []byte, log *zap.SugaredLogger) error {
	client := &http.Client{}
	req, _ := http.NewRequest("POST", c.ApiUrl, bytes.NewBuffer(bodyBytes))
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %v", c.ApiToken))
	req.Header.Add("Content-Type", "application/json")

	response, err := client.Do(req)

	err = checkWebexHttpResponse(response, err)
	if err != nil {
		log.Infof("Failed webex call body contents", "bodyContents", string(bodyBytes))
	}
	if err != nil && strings.Contains(err.Error(), "invalid character") {
		log.Infof("Failed webex call body contents", "bodyContents", string(bodyBytes))
	}
	return err
}
