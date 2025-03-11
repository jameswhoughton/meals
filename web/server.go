package web

import (
	"fmt"
	"net/http"

	"github.com/jameswhoughton/meals/internal/auth"
)

type Server struct {
	port        string
	userService *auth.UserService
}

func (s *Server) Start() error {
	mux := http.NewServeMux()

	AddRoutes(mux, *s.userService)

	fmt.Printf("listening on port :%s\n", s.port)

	return http.ListenAndServe(":"+s.port, mux)
}

func NewServer(port string, userService *auth.UserService) *Server {
	return &Server{
		port:        port,
		userService: userService,
	}
}
