package response

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	apperrors "github.com/ileego/go_react_ai/pkg/errors"
)

func TestFromError_Validation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	FromError(c, apperrors.NewValidation("title", "标题不能为空"))

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
	var body Body[any]
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal body: %v", err)
	}
	if body.ErrCode != "VALIDATION_ERROR" {
		t.Errorf("err_code = %q, want VALIDATION_ERROR", body.ErrCode)
	}
	if body.Message != "标题不能为空" {
		t.Errorf("message = %q, want 标题不能为空", body.Message)
	}
}

func TestFromError_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	FromError(c, apperrors.NewNotFound("report", 42))

	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusNotFound)
	}
	var body Body[any]
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal body: %v", err)
	}
	if body.ErrCode != "NOT_FOUND" {
		t.Errorf("err_code = %q, want NOT_FOUND", body.ErrCode)
	}
}

func TestFromError_CustomCode(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	err := apperrors.NewValidation("status", "当前状态不允许取消").WithCode("REPORT_CANNOT_CANCEL")
	FromError(c, err)

	var body Body[any]
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal body: %v", err)
	}
	if body.ErrCode != "REPORT_CANNOT_CANCEL" {
		t.Errorf("err_code = %q, want REPORT_CANNOT_CANCEL", body.ErrCode)
	}
}

func TestFromError_NonAppError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	FromError(c, errors.New("boom"))

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusInternalServerError)
	}
	var body Body[any]
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal body: %v", err)
	}
	if body.ErrCode != "INTERNAL_ERROR" {
		t.Errorf("err_code = %q, want INTERNAL_ERROR", body.ErrCode)
	}
}
