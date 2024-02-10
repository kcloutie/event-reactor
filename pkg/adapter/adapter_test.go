package adapter

import (
	"testing"

	"github.com/kcloutie/event-reactor/pkg/config"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

func TestGetReactors(t *testing.T) {
	testLogger := zaptest.NewLogger(t)
	reactorConfig := config.ReactorConfig{}

	tests := []struct {
		name        string
		reactorType string
		wantExists  bool
	}{
		{
			name:        "Reactor type is not log",
			reactorType: "not_log",
			wantExists:  false,
		},
		{
			name:        "Reactor type is github/comment",
			reactorType: "github/comment",
			wantExists:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reactorFunctions := GetReactorNewFunctions(false)
			pFunc, exists := reactorFunctions[tt.reactorType]
			assert.Equal(t, tt.wantExists, exists)
			if !exists {
				return
			}
			reactor := pFunc(testLogger, reactorConfig)
			assert.NotNil(t, reactor)
			assert.Equal(t, tt.reactorType, reactor.GetName())
		})
	}
}
