package test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	"cloud.google.com/go/pubsub"
	"github.com/kcloutie/event-reactor/pkg/config"
	"gotest.tools/assert"
)

func TestReactorLog(t *testing.T) {
	if !ShouldRunIntegrationTest(t) {
		return
	}
	testUrl := GetTestUrl()
	postBody := pubsub.Message{
		Attributes: map[string]string{
			"att1": "att1Val",
		},
		Data: []byte(`{"prop1":"123"}`),
		ID:   "1",
	}
	postBodyBytes, err := json.Marshal(postBody)
	assert.NilError(t, err)
	response, err := http.Post(fmt.Sprintf("%s/api/v1/pubsub", testUrl), "application/json", strings.NewReader(string(postBodyBytes)))
	assert.NilError(t, err)
	defer response.Body.Close()
	body, _ := io.ReadAll(response.Body)
	// assert.NilError(t, fmt.Errorf("ttt"))
	assert.Equal(t, response.StatusCode, 200)
	assert.Equal(t, string(body), "")
}

func TestReactorLog2(t *testing.T) {
	ghTemplateBytes, _ := os.ReadFile("testdata/github-comment.md")
	configBytes, _ := os.ReadFile("testdata/testConfig.json")
	sc := config.ServerConfiguration{}
	json.Unmarshal(configBytes, &sc)

	valProp := string(ghTemplateBytes)
	sc.ReactorConfigs[0].Properties["message"] = config.PropertyAndValue{
		Value: &valProp,
	}

	jsonBytes, _ := json.Marshal(sc)
	os.WriteFile("testdata/testConfig2.json", jsonBytes, 0644)

	payloadBytes, _ := os.ReadFile("testdata/data-example.json")
	if !ShouldRunIntegrationTest(t) {
		return
	}
	testUrl := GetTestUrl()
	postBody := pubsub.Message{
		Attributes: map[string]string{
			"test": "github",
		},
		Data: payloadBytes,
		ID:   "1",
	}
	postBodyBytes, err := json.Marshal(postBody)
	assert.NilError(t, err)
	response, err := http.Post(fmt.Sprintf("%s/api/v1/pubsub", testUrl), "application/json", strings.NewReader(string(postBodyBytes)))
	assert.NilError(t, err)
	defer response.Body.Close()
	body, _ := io.ReadAll(response.Body)
	// assert.NilError(t, fmt.Errorf("ttt"))
	assert.Equal(t, response.StatusCode, 200)
	assert.Equal(t, string(body), "")
}
