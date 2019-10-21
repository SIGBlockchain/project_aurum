package requests

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/SIGBlockchain/project_aurum/internal/block"
	"github.com/SIGBlockchain/project_aurum/internal/contracts"
	"github.com/google/go-cmp/cmp"
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

func TestNewContractRequest(t *testing.T) {
	senderPrivateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	testContract, err := contracts.New(1, senderPrivateKey, []byte{1}, 25, 20)
	if err != nil {
		t.Errorf("failed to make contract : %v", err)
	}
	testContract.Sign(senderPrivateKey)
	req, err := NewContractRequest("", *testContract)
	if err != nil {
		t.Errorf("failed to create test contract: %v", err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		buf := new(bytes.Buffer)
		buf.ReadFrom(r.Body)
		io.WriteString(w, buf.String())
	})
	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		t.Logf("%s", rr.Body.String())
	}
	var responseBody contracts.JSONContract
	if err := json.Unmarshal(rr.Body.Bytes(), &responseBody); err != nil {
		t.Errorf("failed to unmarshall response body: %v", err)
	}
	responseContract, err := responseBody.Unmarshal()
	if err != nil {
		t.Errorf("failed to convert JSONContract to Contract: %v", err)
	}
	if !testContract.Equals(responseContract) {
		t.Errorf("contracts do not match:\n got %+v want %+v", responseContract, *testContract)
	}
}

func TestAddPeerToDiscoveryRequest(t *testing.T) {
	// Arrange
	ip := "1.2.3.4"
	port := "9001"
	rr := httptest.NewRecorder()
	req, err := AddPeerToDiscoveryRequest(ip, port)
	if err != nil {
		t.Errorf("failed to create add peer request %v", err)
	}
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, r.URL.Query().Get("ip")+":"+r.URL.Query().Get("port"))
	})

	// Act
	handler.ServeHTTP(rr, req)

	// Assert
	if status := rr.Code; http.StatusOK != status {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		t.Logf("%s", rr.Body.String())
	}
	expected := ip + ":" + port
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
	}
}

func TestGetBlockByHeightRequest(t *testing.T) {
	req, err := GetBlockByHeightRequest(9001)
	if err != nil {
		t.Errorf(err.Error())
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"received": "`+r.URL.Query().Get("h")+`"}`)
	})

	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
	expected := `{"received": "9001"}`
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
	}
}

func TestGetBlockByHashRequest(t *testing.T) {
	req, err := GetBlockByHashRequest("nastyHash")
	if err != nil {
		t.Errorf(err.Error())
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"received": "`+r.URL.Query().Get("p")+`"}`)
	})
	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
	expected := `{"received": "nastyHash"}`
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
	}
}

func TestSendBlockRequest(t *testing.T) {
	testBlock := block.Block{
		Version:        5,
		Height:         3340,
		PreviousHash:   []byte("guavapineapplemango1234567890abc"),
		MerkleRootHash: []byte("grapewatermelonaurum1emonsabcd"),
		Timestamp:      time.Now().UnixNano(),
		Data:           [][]byte{{12, 3}, {132, 90, 23}, {23}},
	}
	testBlock.DataLen = uint16(len(testBlock.Data))
	expectedJsonBlock, err := testBlock.Marshal()
	if err != nil {
		t.Errorf("Failed to marshal a block: %s", err.Error())
	}

	req, err := SendBlockRequest(&testBlock)
	if err != nil {
		t.Errorf("Failed to create send block request: %s", err.Error())
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected a POST request. Found a %s", r.Method)
		}
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			t.Errorf("Failed to create send block request: %s", err.Error())
		}

		actualJsonBlock := block.JSONBlock{}
		err = json.Unmarshal(body, &actualJsonBlock)
		if err != nil {
			t.Errorf("Failed to unmarhsall request body: %s", err.Error())
		}
		if cmp.Equal(actualJsonBlock, expectedJsonBlock) {
			t.Errorf("Failed to get correct JSON block in request.\nGot: %v\nExpected: %v", actualJsonBlock, expectedJsonBlock)
		}

		w.WriteHeader(http.StatusOK)
	})

	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
}
