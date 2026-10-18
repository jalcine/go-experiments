package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)

type Post struct {
	ID     *int   `json:"id,omitempty"`
	UserID int    `json:"userId,omitempty"`
	Title  string `json:"title,omitempty"`
	Body   string `json:"body,omitempty"`
}

type Company struct {
	Name string `json:"name"`
}

type Coordinate struct {
	Lat float32 `json:"lat,string"`
	Lng float32 `json:"lng,string"`
}

type Location struct {
	Street     string     `json:"street"`
	City       string     `json:"city"`
	Coordinate Coordinate `json:"geo"`
}

type Contact struct {
	ID      int      `json:"id"`
	Name    string   `json:"name"`
	Company Company  `json:"company"`
	Address Location `json:"address"`
}

type Comment struct {
	ID    int    `json:"id"`
	Email string `json:"email"`
	Body  string `json:"body"`
	Name  string `json:"name"`
}

func main() {
	fmt.Println("Exercise: ======== printing posts ====")
	client := http.Client{Timeout: 10 * time.Second}
	if err := handle_posts(client); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	fmt.Println("======================================")

	fmt.Println("Exercise: ==== printing users ========")
	if err := handle_users(client); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(2)
	}
	fmt.Println("======================================")
	fmt.Println("Exercise: ==== handling query ========")

	if err := handle_post_and_comments(client); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(3)
	}
	fmt.Println("======================================")
	fmt.Println("Exercise: ==== handling errors ========")

	if err := handle_errors(client); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
	}
	fmt.Println("======================================")
	fmt.Println("Exercise: ==== handling creation =====")

	if err := handle_create(client); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
	}
	fmt.Println("======================================")
	fmt.Println("Exercise: ==== handling updating =====")

	if err := handle_put(client); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
	}
	fmt.Println("======================================")

}

func getJSON[T any](client http.Client, url string, out *T) error {
	resp, err := client.Get(url)

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("request failed: %d", resp.StatusCode)
	}

	return json.NewDecoder(resp.Body).Decode(out)
}

func handle_post_and_comments(client http.Client) error {
	postId := 1
	post_url := "https://jsonplaceholder.typicode.com/posts/" + strconv.Itoa(postId)

	var post Post
	if err := getJSON(client, post_url, &post); err != nil {
		return fmt.Errorf("JSON decoding of post #%d body failed: %w", postId, err)
	}

	fmt.Printf("Post #%d: %s\n", *post.ID, post.Title)

	base_comments_url := "https://jsonplaceholder.typicode.com/comments"
	params := url.Values{}
	params.Set("postId", strconv.Itoa(postId))
	comments_url := base_comments_url + "?" + params.Encode()

	var comments []Comment

	if err := getJSON(client, comments_url, &comments); err != nil {
		return fmt.Errorf("JSON decoding of comments for post #%d: %w", postId, err)
	}

	for _, comment := range comments {
		fmt.Printf("Comment #%d by %s (%s): %s\n", comment.ID, comment.Name, comment.Email, comment.Body)
	}

	return nil
}

func handle_users(client http.Client) error {

	var users []Contact
	if err := getJSON(client, "https://jsonplaceholder.typicode.com/users", &users); err != nil {
		return fmt.Errorf("decoding users: %w", err)
	}

	user := users[0]

	fmt.Printf("User %s from %s (%s)\n", user.Name, user.Address.City, user.Company.Name)

	return nil
}

func handle_posts(client http.Client) error {
	var posts []Post
	if err := getJSON(client, "https://jsonplaceholder.typicode.com/posts", &posts); err != nil {
		return fmt.Errorf("decoding posts: %w", err)
	}

	fmt.Println("size", len(posts))

	if len(posts) == 0 {
		return fmt.Errorf("no posts returned")
	}

	post := posts[0]
	fmt.Printf("Post %d: %s\n", *post.ID, post.Title)
	return nil
}

type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func handle_errors(client http.Client) error {
	postId := 9999
	postUrl := "https://jsonplaceholder.typicode.com/posts/" + strconv.Itoa(postId)

	resp, err := client.Get(postUrl)

	if err != nil {
		return fmt.Errorf("fetching post #%d: %w", postId, err)
	}

	defer resp.Body.Close()

	var post Post

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return fmt.Errorf("failed to read JSON body from response: %w", err)
	}

	if err := json.Unmarshal(body, &post); err != nil {
		return fmt.Errorf("failed to unmarshal JSON response: %w,", err)
	}

	if post.ID == nil {
		return fmt.Errorf("the post was invalid (status %d, body %s)", resp.StatusCode, body)
	}

	fmt.Printf("Post title %s\n", post.Title)

	return nil
}

func handle_create(client http.Client) error {
	newPost := Post{UserID: 1, Title: "A New Post", Body: "A New Body"}
	json_body, err := json.Marshal(&newPost)

	if err != nil {
		return fmt.Errorf("failed to marshal post body: %w", err)
	}

	json_bytes := bytes.NewReader(json_body)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", "https://jsonplaceholder.typicode.com/posts", json_bytes)

	if err != nil {
		return fmt.Errorf("failed to form request to make post: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)

	if err != nil {
		return fmt.Errorf("failed to complete request for post creation: %w", err)
	}

	defer resp.Body.Close()

	var post Post

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return fmt.Errorf("failed to read body from response: %w", err)
	}

	if resp.StatusCode != 201 {
		return fmt.Errorf("failed to create a new post: %s", body)
	}

	if err := json.Unmarshal(body, &post); err != nil {
		return fmt.Errorf("failed to decode the JSON as a post: %w", err)
	}

	fmt.Printf("created post id=%d title=%s\n", *post.ID, post.Title)

	return nil
}

type HttpClient struct {
	BaseURL string
	Client  *http.Client
}

func (client *HttpClient) ReplacePost(ctx context.Context, post Post) (*Post, error) {
	post_url := client.BaseURL + "/posts/" + strconv.Itoa(*post.ID)

	json_string, err := json.Marshal(post)

	if err != nil {
		return nil, fmt.Errorf("failed to marshal into JSON: %w", err)
	}

	json_bytes := bytes.NewReader(json_string)

	req, err := http.NewRequestWithContext(ctx, "PUT", post_url, json_bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to form update request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Client.Do(req)

	if err != nil {
		return nil, fmt.Errorf("failed to complete update request: %w", err)
	}

	defer resp.Body.Close()

	var parsed_post Post

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return nil, fmt.Errorf("failed to read body from response: %w", err)
	}

	if err := json.Unmarshal(body, &parsed_post); err != nil {
		return nil, fmt.Errorf("failed to decode the JSON as a post: %w", err)
	}

	return &parsed_post, nil
}

func (client *HttpClient) PatchPost(ctx context.Context, post PostPatch) (*Post, error) {
	post_url := client.BaseURL + "/posts/" + strconv.Itoa(post.ID)

	json_string, err := json.Marshal(post)

	if err != nil {
		return nil, fmt.Errorf("failed to marshal into JSON: %w", err)
	}

	json_bytes := bytes.NewReader(json_string)

	req, err := http.NewRequestWithContext(ctx, "PATCH", post_url, json_bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to form patch request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Client.Do(req)

	if err != nil {
		return nil, fmt.Errorf("failed to complete patch request: %w", err)
	}

	defer resp.Body.Close()

	var parsed_post Post

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return nil, fmt.Errorf("failed to read body from response: %w", err)
	}

	if err := json.Unmarshal(body, &parsed_post); err != nil {
		return nil, fmt.Errorf("failed to decode the JSON as a post: %w", err)
	}

	return &parsed_post, nil
}

type PostPatch struct {
	Title *string `json:"title,omitempty"`
	Body  *string `json:"body,omitempty"`
}

func handle_put(client http.Client) error {
	api := &HttpClient{
		BaseURL: "https://jsonplaceholder.typicode.com",
		Client:  &client,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()

	updated, err := api.ReplacePost(ctx, Post{
		ID: new(1), UserID: 1, Title: "updated", Body: "replaced Body",
	})

	if err != nil {
		return err
	}

	fmt.Printf("replaced post: id=%d title=%s\n", *updated.ID, updated.Title)

	patched, err := api.PatchPost(ctx, PostPatch{
		Title: new("patched"),
	})
	if err != nil {
		return err
	}

	fmt.Printf("patched post: id=%d title=%s  body=%s\n", *patched.ID, patched.Title, patched.Body)

	return nil
}
