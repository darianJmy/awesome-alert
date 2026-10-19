package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"aiops-backend/models"
	"github.com/gin-gonic/gin"
)

func TestRetiredProbeTypesRejected(t *testing.T) {
	for _, probeType := range []string{"mqtt", "ssl"} {
		t.Run(probeType, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(recorder)
			ctx.Request = httptest.NewRequest(http.MethodPost, "/api/instant-probe", strings.NewReader(`{"type":"`+probeType+`","host":"127.0.0.1"}`))
			ctx.Request.Header.Set("Content-Type", "application/json")
			InstantProbe(ctx)
			if recorder.Code != http.StatusBadRequest {
				t.Fatalf("retired instant probe accepted: %d", recorder.Code)
			}
			if err := validateInspectionItemReq(CreateInspectionItemReq{Type: probeType}); err == nil || !strings.Contains(err.Error(), "不支持的拨测类型") {
				t.Fatalf("retired task configuration accepted: %v", err)
			}
			status, _, _ := executeInspection(models.InspectionItem{Type: probeType})
			if status != "error" {
				t.Fatalf("retired task executed: %s", status)
			}
		})
	}
}
