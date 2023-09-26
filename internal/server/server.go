package server

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/zhigulin-io/word-stat/internal/entity"
)

type StatsProvider interface {
	GetStatsForPost(postID int) []entity.Stat
}

type Server struct {
	config        Config
	statsProvider StatsProvider
}

func New(config Config, statsProvider StatsProvider) *Server {
	return &Server{
		config:        config,
		statsProvider: statsProvider,
	}
}

func (s Server) Serve() {
	r := gin.Default()

	r.GET("/post/:id/comments/statistics", func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"err": "cannot convert id to int",
			})
			return
		}

		stats := s.statsProvider.GetStatsForPost(id)

		var status int
		if len(stats) != 0 {
			status = http.StatusOK
		} else {
			status = http.StatusNotFound
		}

		c.JSON(status, stats)
	})

	r.Run(fmt.Sprintf("%s:%d", s.config.Host, s.config.Port))
}
