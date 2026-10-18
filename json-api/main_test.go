package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"
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

func TestHttpClient_DeletePost_204(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	ts := httptest.NewServer(handler)
	defer ts.Close()

	client := &HttpClient{
		BaseURL: ts.URL,
		Client:  &http.Client{},
	}

	t.Run("deletes a post", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := client.DeletePost(ctx, 1); err != nil {
			t.Fatalf("expected no error on deletion; got %v", err)
		}

	})
}

func TestHttpClient_DeletePost_200(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	})
	ts := httptest.NewServer(handler)
	defer ts.Close()

	client := &HttpClient{
		BaseURL: ts.URL,
		Client:  &http.Client{},
	}

	t.Run("deletes a post", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := client.DeletePost(ctx, 1); err != nil {
			t.Fatalf("expected no error on deletion; got %v", err)
		}

	})
}

func TestHttpClient_DeletePost_500(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message": "Failed to work."}`))
	})
	ts := httptest.NewServer(handler)
	defer ts.Close()

	client := &HttpClient{
		BaseURL: ts.URL,
		Client:  &http.Client{},
	}

	t.Run("deletes a post", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := client.DeletePost(ctx, 1); err == nil {
			t.Fatalf("expected an error on deletion; got nothing")
		}

	})
}

func postGen() func() Post {
	i := 0

	return func() Post {
		i += 1

		return Post{
			ID:    &FlexibleID{ID: i},
			Title: "Post " + strconv.Itoa(i),
			Body:  "The contents of post #" + strconv.Itoa(i),
		}

	}

}

func TestHttpClient_Pagination_MultiPage(t *testing.T) {
	nextPost := postGen()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var posts []Post

		page_num := r.URL.Query().Get("_page")
		limit, _ := strconv.Atoi(r.URL.Query().Get("_limit"))

		switch page_num {
		case "1", "2", "3":
			{
				for range limit {
					posts = append(posts, nextPost())
				}
			}
		case "4":
			posts = []Post{nextPost()}
		}

		json_bytes, err := json.Marshal(posts)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
		} else {
			w.Write(json_bytes)
		}
	})
	ts := httptest.NewServer(handler)
	defer ts.Close()

	client := &HttpClient{
		BaseURL: ts.URL,
		Client:  &http.Client{},
	}

	t.Run("paginates over all posts", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		posts, err := client.ListAllPosts(ctx)

		if err != nil {
			t.Fatalf("expected no error on pagination; got %v", err)
		}

		if len(posts) == 0 {
			t.Fatalf("no posts were retrieved")
		}

	})
}
