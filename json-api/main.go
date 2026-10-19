package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type FlexibleID struct {
	ID int
}

func (f *FlexibleID) UnmarshalJSON(data []byte) error {
	if data[0] == '"' || data[0] == 't' || data[0] == 'f' {
		var s string
		if err := json.Unmarshal(data, &s); err != nil {
			return err
		}
		num, err := strconv.Atoi(s)
		if err != nil {
			return err
		}
		f.ID = num
	} else {
		var num int
		if err := json.Unmarshal(data, &num); err != nil {
			return err
		}
		f.ID = num
	}

	return nil
}

func (f FlexibleID) MarshalJSON() ([]byte, error) {
	return json.Marshal(f.ID)
}

type Post struct {
	ID     *FlexibleID `json:"id,omitempty"`
	UserID int         `json:"userId,omitempty"`
	Title  string      `json:"title,omitempty"`
	Body   string      `json:"body,omitempty"`
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
	fmt.Println("Exercise: ==== handling concurrency =====")

	if err := handle_user_posts(client); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
	}
	fmt.Println("======================================")

}

var ErrBadStatus = errors.New("bad status code")

func getJSON[T any](client http.Client, url string, out *T) error {
	resp, err := client.Get(url)

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("request failed: %d: %w", resp.StatusCode, ErrBadStatus)
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

func (client *HttpClient) ListAllPosts(ctx context.Context) ([]Post, error) {
	var posts []Post
	page_num := 1
	limit := 10

	for {
		current_page_url := client.BaseURL + "/posts?_page=" + strconv.Itoa(page_num) + "&_limit=" + strconv.Itoa(limit)
		req, err := http.NewRequestWithContext(ctx, "GET", current_page_url, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to form pagination request for page %d: %v", page_num, err)
		}

		resp, err := client.Client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("pagination of page %d request failed: %v", page_num, err)
		}

		defer resp.Body.Close()

		bytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read body from response for page %d: %v", page_num, err)
		}

		if resp.StatusCode >= 500 {
			return nil, fmt.Errorf("failed to fetch more posts on page #%d, %s", page_num, bytes)
		}

		var page_posts []Post
		if err := json.Unmarshal(bytes, &page_posts); err != nil {
			return nil, fmt.Errorf("failed to decode JSON for page %d: %v", page_num, err)
		}

		if len(page_posts) == 0 {
			break
		}

		posts = append(posts, page_posts...)
		page_num++
	}

	return posts, nil
}

func (client *HttpClient) DeletePost(ctx context.Context, postId int) error {
	post_url := client.BaseURL + "/posts/" + strconv.Itoa(postId)

	req, err := http.NewRequestWithContext(ctx, "DELETE", post_url, nil)
	if err != nil {
		return fmt.Errorf("failed to form deletion request: %v", err)
	}

	resp, err := client.Client.Do(req)
	if err != nil {
		return fmt.Errorf("deletion request failed: %v", err)
	}

	defer resp.Body.Close()

	bytes, err := io.ReadAll(resp.Body)

	if err != nil {
		return fmt.Errorf("failed to parse response body: %v", err)
	}

	switch resp.StatusCode {
	case 204:
		{
			if resp.ContentLength != 0 {
				return fmt.Errorf("did not expect a body")
			}
		}
	case 200:
		{
			var obj = map[string]any{}
			if err = json.Unmarshal(bytes, &obj); err != nil {
				return fmt.Errorf("failed to convert body into JSON (body %v): %v", bytes, err)
			}

			if len(obj) != 0 {
				return fmt.Errorf("expected empty object as response")
			}
		}
	case 500:
		{
			var obj = map[string]any{}
			if err = json.Unmarshal(bytes, &obj); err != nil {
				return fmt.Errorf("failed to convert body into JSON (body %v): %v", bytes, err)
			}

			return fmt.Errorf("got 500: %v", obj)
		}

	}

	return nil

}

func (client *HttpClient) ReplacePost(ctx context.Context, post Post) (*Post, error) {
	post_url := client.BaseURL + "/posts/" + strconv.Itoa(post.ID.ID)

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

func (client *HttpClient) CreatePostForm(ctx context.Context, post Post) error {
	var values = url.Values{}
	values.Add("title", post.Title)
	values.Add("id", strconv.Itoa(post.ID.ID))
	values.Add("user_id", strconv.Itoa(post.UserID))
	values.Add("body", post.Body)
	form_body := values.Encode()
	form_reader := strings.NewReader(form_body)

	create_post_url := client.BaseURL + "/posts"
	req, err := http.NewRequestWithContext(ctx, "POST", create_post_url, form_reader)

	if err != nil {
		return fmt.Errorf("failed to form request to make post: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send form request for post: %w", err)
	}

	content_type := resp.Header.Get("Content-Type")
	if strings.HasPrefix(content_type, "application/json") {
		return fmt.Errorf("invalid content type in response; expected json, got %s", content_type)
	}

	defer resp.Body.Close()

	var new_post Post

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("the post was not created: %s", ErrBadStatus)
	}

	if err = json.NewDecoder(resp.Body).Decode(&new_post); err != nil {
		return fmt.Errorf("failed to parse JSON into post object %s", err)
	}

	return nil

}

type PostPatch struct {
	ID    int
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
		ID: new(FlexibleID{}), UserID: 1, Title: "updated", Body: "replaced Body",
	})

	if err != nil {
		return err
	}

	fmt.Printf("replaced post: id=%d title=%s\n", *updated.ID, updated.Title)

	patched, err := api.PatchPost(ctx, PostPatch{
		ID:    updated.ID.ID,
		Title: new("patched"),
	})
	if err != nil {
		return err
	}

	fmt.Printf("patched post: id=%d title=%s  body=%s\n", *patched.ID, patched.Title, patched.Body)

	return nil
}

func (client *HttpClient) GetUserPosts(ctx context.Context, userId int) error {
	user_posts_url := client.BaseURL + "/users/" + strconv.Itoa(userId) + "/posts"

	req, err := http.NewRequestWithContext(ctx, "GET", user_posts_url, nil)

	if err != nil {
		return fmt.Errorf("failed to build request: %w", err)
	}

	resp, err := client.Client.Do(req)

	if err != nil {
		return fmt.Errorf("failed to invoke request to stream posts: %w", err)
	}

	var postIds []int

	defer resp.Body.Close()

	dec := json.NewDecoder(resp.Body)

	first_char, err := dec.Token()
	if err != nil {
		return fmt.Errorf("failed to pluck first token: %s", err)
	}

	if first_char != json.Delim('[') {
		return fmt.Errorf("the opening character was not [, it was %s", first_char)
	}

	for dec.More() {
		var p Post

		if err := dec.Decode(&p); err != nil {
			return err
		}

		fmt.Println(*p.ID, p.Title)
		postIds = append(postIds, p.ID.ID)
	}

	counts := make([]int, len(postIds))
	comment_errors := make([]error, len(postIds))

	var wg sync.WaitGroup

	for i, id := range postIds {
		wg.Add(1)
		go func() {
			defer wg.Done()
			comments, err := fetch_comments(client, id)
			if err != nil {
				comment_errors[i] = fmt.Errorf("failed to get comments for post #%d", id)
				return
			}

			counts[i] = len(comments)
		}()

	}

	wg.Wait()

	for i, id := range postIds {
		if comment_errors[i] != nil {
			fmt.Fprintf(os.Stderr, "post %d: %v\n", id, comment_errors[i])
		} else {
			fmt.Printf("post %d: %d comments\n", id, counts[i])
		}
	}

	return nil
}

func fetch_comments(client *HttpClient, postId int) ([]Comment, error) {
	var comments []Comment
	comments_url := client.BaseURL + "/posts/" + strconv.Itoa(postId) + "/comments"
	if err := getJSON(*client.Client, comments_url, &comments); err != nil {
		return nil, err
	}
	return comments, nil
}

func handle_user_posts(client http.Client) error {
	api := &HttpClient{
		BaseURL: "https://jsonplaceholder.typicode.com",
		Client:  &client,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()

	return api.GetUserPosts(ctx, 1)
}
