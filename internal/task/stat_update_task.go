package task

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/zhigulin-io/word-stat/internal/entity"
)

type Comment struct {
	Body string `json:"body"`
}

type Post struct {
	ID int `json:"id"`
}

type Storage interface {
	UpdateStatsForPost(int, []entity.Stat)
}

type StatUpdateTask struct {
	Storage Storage
}

func (s StatUpdateTask) Action() {
	log.Println("StatUpdateTask started")
	resp, err := http.Get("http://jsonplaceholder.typicode.com/posts")
	if err != nil {
		log.Println(err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return
	}

	posts := make([]Post, 0, 100)

	err = json.Unmarshal(body, &posts)
	if err != nil {
		log.Println(err)
		return
	}

	for _, v := range posts {
		go s.handlePost(v.ID)
	}
}

func (s StatUpdateTask) handlePost(postId int) {
	url := fmt.Sprintf(
		"http://jsonplaceholder.typicode.com/comments?postId=%d",
		postId,
	)

	resp, err := http.Get(url)
	if err != nil {
		log.Println(err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return
	}

	comments := make([]Comment, 0, 10)

	err = json.Unmarshal(body, &comments)
	if err != nil {
		log.Println(err)
		return
	}

	go s.handleComments(postId, comments)
}

func (s StatUpdateTask) handleComments(postId int, comments []Comment) {
	if len(comments) == 0 {
		return
	}

	stats := make(map[string]entity.Stat)

	for _, c := range comments {
		words := strings.Fields(c.Body)
		for _, w := range words {
			stat, ok := stats[w]
			if ok {
				stat.Count += 1
				stats[w] = stat
			} else {
				stats[w] = entity.Stat{
					PostID: postId,
					Word:   w,
					Count:  1,
				}
			}
		}
	}

	sliceOfStats := make([]entity.Stat, 0, len(stats))

	for _, stat := range stats {
		sliceOfStats = append(sliceOfStats, stat)
	}

	go s.Storage.UpdateStatsForPost(postId, sliceOfStats)
}
