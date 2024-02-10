package http

import (
	"context"
	"reflect"
	"testing"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func TestSetCommonLoggingAttributes(t *testing.T) {
	type args struct {
		ctx context.Context
		c   *gin.Context
	}
	tests := []struct {
		name  string
		args  args
		want  *zap.Logger
		want1 context.Context
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := SetCommonLoggingAttributes(tt.args.ctx, tt.args.c)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SetCommonLoggingAttributes() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("SetCommonLoggingAttributes() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
