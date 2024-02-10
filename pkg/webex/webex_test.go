package webex

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap/zaptest"
)

func TestSendWithCard(t *testing.T) {
	testLogger := zaptest.NewLogger(t)
	spaceId := "123"
	apiToken := "token"
	message := "message"
	card := `{"$schema":"http://adaptivecards.io/schemas/adaptive-card.json","type":"AdaptiveCard","version":"1.2","body":[{"type":"ColumnSet","columns":[{"type":"Column","width":"stretch","items":[{"type":"TextBlock","text":"TIMEOUT","wrap":true,"size":"large","weight":"bolder","color":"accent","horizontalAlignment":"left","id":"PIPELINE_NAME"}],"id":"PIPELINE_NAME_COL"},{"type":"Column","width":"auto","items":[{"type":"TextBlock","text":"Failed","wrap":true,"size":"large","weight":"bolder","color":"good","horizontalAlignment":"right","isSubtle":false,"id":"PIPELINE_STATUS"}],"id":"PIPELINE_STATUS_COL"}],"id":"HEADING_COL_SET"},{"type":"Container","items":[{"type":"ColumnSet","style":"default","columns":[{"type":"Column","width":"stretch","items":[{"type":"TextBlock","text":"Pipeline Run:","wrap":true,"size":"Medium","weight":"bolder","id":"PIPELINE_RUN_LAB"}],"id":"PIPELINE_RUN_COL"},{"type":"Column","width":"stretch","items":[{"type":"Container","selectAction":{"type":"Action.OpenUrl","url":"https://console.cloud.google.com/storage/browser/_details/gscBucket/gscBucket/path","title":"timeout-w426s7"},"items":[{"type":"TextBlock","wrap":true,"color":"accent","fontType":"Default","isSubtle":true,"text":"timeout-w426s7","id":"PIPELINE_RUN_NAME"}],"id":"PIPELINE_RUN_NAME_CONT"}],"id":"PIPELINE_RUN_NAME_COL"}],"id":"PIPELINE_RUN_COL_SET"},{"type":"ColumnSet","columns":[{"type":"Column","width":"stretch","items":[{"type":"TextBlock","text":"Pipeline Run Time:","wrap":true,"weight":"bolder","size":"Medium","id":"PIPELINE_RUN_TIME_LAB"}]},{"type":"Column","width":"stretch","items":[{"type":"TextBlock","text":"40s","wrap":true,"id":"PIPELINE_RUNTIME"}]}]},{"type":"ColumnSet","columns":[{"type":"Column","width":"stretch","items":[{"type":"TextBlock","text":"Failed Tasks:","wrap":true,"weight":"bolder","size":"Medium"}]},{"type":"Column","width":"stretch","items":[{"type":"TextBlock","text":"sleep-task","wrap":true,"id":"FAILED_TASKS"}]}]},{"type":"ColumnSet","columns":[{"type":"Column","width":"stretch","items":[{"type":"TextBlock","text":"Result Details:","wrap":true,"weight":"bolder","size":"Medium"}]},{"type":"Column","width":"stretch","items":[{"type":"TextBlock","text":"PipelineRunTimeout - PipelineRun timeout-w426s7 failed to finish within 40s","wrap":true,"id":"STATUS_DETAILS"}]}]}],"separator":true,"id":"PIPELINE_CONT"}]}`
	badCard := "card"
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/Success" {
			hadError := false
			if req.Header["Authorization"][0] != fmt.Sprintf("Bearer %v", apiToken) {
				hadError = true
				_, _ = rw.Write([]byte("Invalid token"))
			}

			bodyObject := MessageCreateRequest{}
			buf := new(bytes.Buffer)
			buf.ReadFrom(req.Body)
			reqBody := buf.String()
			unmarshalErr := json.Unmarshal([]byte(reqBody), &bodyObject)
			if unmarshalErr != nil {
				hadError = true
				_, _ = rw.Write([]byte("Failed to unmarshal body"))
			}

			if bodyObject.RoomID != spaceId {
				hadError = true
				_, _ = rw.Write([]byte("Invalid spaceId"))
			}

			if bodyObject.Text != message {
				hadError = true
				_, _ = rw.Write([]byte("Invalid message"))
			}

			if hadError {
				rw.WriteHeader(500)
			} else {
				rw.WriteHeader(200)
			}
		}
	}))
	defer server.Close()

	type args struct {
		config WebexConfiguration
	}
	tests := []struct {
		name             string
		args             args
		wantResponseCode int
		wantBody         string
		wantErr          bool
	}{
		{
			name: "Success",
			args: args{
				config: WebexConfiguration{
					Log:      testLogger,
					ApiUrl:   fmt.Sprintf("%v/Success", server.URL),
					ApiToken: apiToken,
					SpaceId:  spaceId,
					Message:  message,
					Card:     card,
				},
			},
			wantErr:          false,
			wantResponseCode: 200,
			wantBody:         "",
		},
		{
			name: "bad_card_contents",
			args: args{
				config: WebexConfiguration{
					Log:      testLogger,
					ApiUrl:   fmt.Sprintf("%v/Success", server.URL),
					ApiToken: apiToken,
					SpaceId:  spaceId,
					Message:  message,
					Card:     badCard,
				},
			},
			wantErr:          true,
			wantResponseCode: 200,
			wantBody:         "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			err := tt.args.config.SendWithCard()
			if (err != nil) != tt.wantErr {
				t.Errorf("SendWithCard() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestSendMessage(t *testing.T) {
	testLogger := zaptest.NewLogger(t)
	spaceId := "123"
	apiToken := "token"
	message := "message"
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/Success" {
			hadError := false
			if req.Header["Authorization"][0] != fmt.Sprintf("Bearer %v", apiToken) {
				hadError = true
				_, _ = rw.Write([]byte("Invalid token"))
			}

			bodyObject := MessageCreateRequest{}
			buf := new(bytes.Buffer)
			buf.ReadFrom(req.Body)
			reqBody := buf.String()
			unmarshalErr := json.Unmarshal([]byte(reqBody), &bodyObject)
			if unmarshalErr != nil {
				hadError = true
				_, _ = rw.Write([]byte("Failed to unmarshal body"))
			}

			if bodyObject.RoomID != spaceId {
				hadError = true
				_, _ = rw.Write([]byte("Invalid spaceId"))
			}

			if bodyObject.Text != message {
				hadError = true
				_, _ = rw.Write([]byte("Invalid message"))
			}

			if hadError {
				rw.WriteHeader(500)
			} else {
				rw.WriteHeader(200)
			}
		}
	}))
	defer server.Close()

	type args struct {
		config WebexConfiguration
	}
	tests := []struct {
		name             string
		args             args
		want             *http.Response
		wantErr          bool
		wantResponseCode int
		wantBody         string
	}{
		{
			name: "Success",
			args: args{
				config: WebexConfiguration{
					Log:      testLogger,
					ApiUrl:   fmt.Sprintf("%v/Success", server.URL),
					ApiToken: apiToken,
					SpaceId:  spaceId,
					Message:  message,
				},
			},
			wantErr:          false,
			wantBody:         "",
			wantResponseCode: 200,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			err := tt.args.config.SendMessage()

			if (err != nil) != tt.wantErr {
				t.Errorf("SendMessage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

		})
	}
}
