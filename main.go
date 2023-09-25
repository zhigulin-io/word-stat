package main

// TODO: Разбить на отдельные файлы

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-co-op/gocron"

	_ "github.com/lib/pq"
)

type Stat struct {
	PostID int    `json:"postId"`
	Word   string `json:"word"`
	Count  int    `json:"count"`
}

type Comment struct {
	Body string `json:"body"`
}

type Post struct {
	ID int `json:"id"`
}

func main() {
	// TODO: Вынести в отдельный сервис или избежать повторения при горизонтальном масштабировании
	scheduler := gocron.NewScheduler(time.UTC)
	scheduler.Cron("*/5 * * * *").Do(updateStat)
	scheduler.StartAsync()

	r := gin.Default()

	r.GET("/post/:id/comments/statistics", func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"err": "cannot convert id to int",
			})
			return
		}

		stats := fetchStatFromDB(id)

		if len(stats) != 0 {
			c.JSON(http.StatusOK, stats)
		} else {
			c.JSON(http.StatusNotFound, gin.H{
				"err": fmt.Sprintf("Stat not found for post with ID: %d", id),
			})
		}
	})

	r.Run()
}

// TODO: достать из базы данных всю статистику для поста в порядке сортировки
func fetchStatFromDB(postId int) []Stat {
	result := make([]Stat, 2)
	result[0] = Stat{
		PostID: postId,
		Word:   "Hello",
		Count:  3,
	}
	result[1] = Stat{
		PostID: postId,
		Word:   "World",
		Count:  5,
	}

	return result
}

func updateStat() {
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
		go handlePost(v.ID)
	}
}

func handlePost(postId int) {
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

	go handleComments(postId, comments)
}

func handleComments(postId int, comments []Comment) {
	if len(comments) == 0 {
		return
	}

	stats := make(map[string]Stat)

	for _, c := range comments {
		words := strings.Fields(c.Body)
		for _, w := range words {
			stat, ok := stats[w]
			if ok {
				stat.Count += 1
				stats[w] = stat
			} else {
				stats[w] = Stat{
					PostID: postId,
					Word:   w,
					Count:  1,
				}
			}
		}
	}

	go writeStatsToDb(stats)
}

// TODO: Сделать пакетную запись в базу данных
func writeStatsToDb(stats map[string]Stat) {
	for _, s := range stats {
		log.Println(s)
	}
}
