package controller

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"mangahub-backend/internal/core/ws"
	"mangahub-backend/internal/modules/auth/service"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// Allow all origins for dev/lab purposes
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type WSHandler struct {
	hub     *ws.Hub
	authSvc *service.AuthService
}

func NewWSHandler(hub *ws.Hub, authSvc *service.AuthService) *WSHandler {
	return &WSHandler{
		hub:     hub,
		authSvc: authSvc,
	}
}

func (h *WSHandler) HandleWS(c *gin.Context) {
	// Extract JWT token from query string ?token= or Sec-WebSocket-Protocol header
	tokenString := c.Query("token")
	if tokenString == "" {
		protocols := c.Request.Header["Sec-Websocket-Protocol"]
		if len(protocols) > 0 {
			tokenString = protocols[0]
		}
	}

	if tokenString == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
		return
	}

	userID, err := h.authSvc.VerifyToken(tokenString)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}

	// Upgrade connection
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("ws upgrade error: %v", err)
		return
	}

	client := ws.NewClient(h.hub, conn, userID)
	h.hub.Register(client) // Note: we need to export Register if not already

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.WritePump()
	go client.ReadPump()
}
