package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFlexibleID_DecodesFromNumberOrString(t *testing.T) {
	cases := []struct {
		name    string
		body    string
		want    int
		wantErr bool
	}{
		{"number", `1`, 1, false},
		{"string", `"1"`, 1, false},
		{"stringErr", `"banana"`, 0, true},
		{"boolean", `true`, 0, true},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			var flexi_id FlexibleID
			err := json.Unmarshal([]byte(testCase.body), &flexi_id)
			if testCase.wantErr == true {
				if err != nil {
					return
				}
			} else {
				if err != nil {
					t.Fatalf("unmarshal failed: %s", err)
				}
			}

			if testCase.want != flexi_id.ID {
				t.Fatalf("wanted %d, obtained %d", testCase.want, flexi_id.ID)
			}
		})
	}
}

func TestPost_DecodesFromEmptyObject(t *testing.T) {
	t.Run("empty object", func(t *testing.T) {
		var p Post
		if err := json.Unmarshal([]byte("{}"), &p); err != nil {
			t.Fatalf("expected empty object; failed with %s", err)
		}

		if p.ID != nil {
			t.Fatalf("the ID was not set to nil on init")
		}
	})
}

func TestGetJson_Decode_200(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id": 1}`))
	})
	ts := httptest.NewServer(handler)
	defer ts.Close()

	var post Post
	err := getJSON(http.Client{}, ts.URL, &post)
	if err != nil {
		t.Fatalf("expected no error, got %s", err)
	}

}

func TestGetJson_Decode_200_Malformed(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id": 1`))
	})
	ts := httptest.NewServer(handler)
	defer ts.Close()

	var post Post
	err := getJSON(http.Client{}, ts.URL, &post)
	if err == nil || errors.Is(err, ErrBadStatus) {
		t.Fatalf("want a non-status (decode) error, got %v", err)
	}

}

func TestGetJson_Decode_500(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"id": 1}`))
	})
	ts := httptest.NewServer(handler)
	defer ts.Close()

	var post Post
	err := getJSON(http.Client{}, ts.URL, &post)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

}
