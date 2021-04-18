package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
)

const (
	POST_BY_ID_URL        = "https://jsonplaceholder.typicode.com/posts/%d"
	POSTS_OF_USER_URL     = "https://jsonplaceholder.typicode.com/posts?userId=%d"
	COMMENTS_FOR_POST_URL = "https://jsonplaceholder.typicode.com/comments?postId=%d"
)

type Post struct {
	UserId int
	Id     int
	Title  string
	Body   string
}

func (p Post) String() string {
	return fmt.Sprintf("UserId: %d\nId: %d\nTitle: %s\nBody: %s\n",
		p.UserId, p.Id, p.Title, p.Body)
}

type Posts []Post

func (posts *Posts) Get(userId int) error {
	data, err := GetData(fmt.Sprintf(POSTS_OF_USER_URL, userId))
	if err != nil {
		return err
	}
	return json.Unmarshal(data, posts)
}

type Comment struct {
	PostId int
	Id     int
	Name   string
	Email  string
	Body   string
}

func (c Comment) String() string {
	return fmt.Sprintf("PostId: %d\nId: %d\n Name: %s\n Email: %s\n Body: %s\n",
		c.PostId, c.Id, c.Name, c.Email, c.Body)
}

type Comments []Comment

func (comments *Comments) Get(postId int) error {
	data, err := GetData(fmt.Sprintf(COMMENTS_FOR_POST_URL, postId))
	if err != nil {
		return err
	}
	return json.Unmarshal(data, comments)
}

func GetData(url string) ([]byte, error) {
	var body []byte
	resp, err := http.Get(url)
	if err != nil {
		return body, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func addPostToDB(post Post, stmt *sql.Stmt) {
	_, err := stmt.Exec(post.UserId, post.Id, post.Title, post.Body)
	if err != nil {
		log.Println(err.Error())
		return
	}
}

func processComments(postId int, stmt *sql.Stmt) {
	comments := new(Comments)
	err := comments.Get(postId)
	if err != nil {
		log.Println(err.Error())
		return
	}
	for _, c := range *comments {
		_, err := stmt.Exec(c.PostId, c.Id, c.Name, c.Email, c.Body)
		if err != nil {
			log.Println(err.Error())
		}
	}
}

func main() {
	db, err := sql.Open("mysql",
		"dev:devGOlang@tcp(127.0.0.1:3306)/test")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	stmtPost, err := db.Prepare("INSERT INTO posts(user_id, id, title, body) VALUES(?, ?, ?, ?)")
	if err != nil {
		log.Fatal(err)
	}
	stmtComment, err := db.Prepare("INSERT INTO comments(post_id, id, name, email, body) VALUES(?, ?, ?, ?, ?)")
	if err != nil {
		log.Fatal(err)
	}

	posts := new(Posts)

	err = posts.Get(7)
	if err != nil {
		log.Fatal(err.Error())
	}
	for _, post := range *posts {
		go addPostToDB(post, stmtPost)
		go processComments(post.Id, stmtComment)
	}

	var input string
	fmt.Scanln(&input)
}
