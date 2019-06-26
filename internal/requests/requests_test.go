package requests

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAccountInfoRequest(t *testing.T) {
	req, err := NewAccountInfoRequest("", "xyz")
	if err != nil {
		t.Errorf("failed to create new account info request")
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"received": "`+r.URL.Query().Get("w")+`"}`)
	})
	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
	expected := `{"received": "xyz"}`
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
	}
}
