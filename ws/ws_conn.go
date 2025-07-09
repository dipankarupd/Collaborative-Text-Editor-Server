package ws

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/dipankarupd/text-editor/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"gorm.io/gorm"
)

var db *gorm.DB
func InitDb(database *gorm.DB) {
	db = database
}

type Message struct {
	Event string          `json:"event"` // "join", "typing", "save"
	Room  string          `json:"room"`  // documentId
	Data  json.RawMessage `json:"data"`  // delta for "typing"
}

type Connection struct {
	ws     *websocket.Conn
	roomID string
}

type RoomManager struct {
	sync.Mutex
	rooms map[string]map[*Connection]bool
}

var manager = RoomManager{rooms: make(map[string]map[*Connection]bool)}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func WebSocketHandler(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}
	
	client := &Connection{ws: conn}

	defer func() {
		removeClientFromRoom(client)
		conn.Close()
	}()

	for {
		var msg Message
		if err := conn.ReadJSON(&msg); err != nil {
			log.Println("WebSocket read error:", err)
			break
		}

		switch msg.Event {
		case "join":
			client.roomID = msg.Room
			addClientToRoom(client)
			log.Printf("Client joined room: %s\n", msg.Room)

		case "typing":
			broadcastToOthers(client, msg)

		case "save":
			saveMessage(msg.Room, msg.Data)

		default:
			log.Printf("Unknown event: %s\n", msg.Event)
		}
	}
}


func addClientToRoom(c *Connection) {
	manager.Lock()
	defer manager.Unlock()

	if manager.rooms[c.roomID] == nil {
		manager.rooms[c.roomID] = make(map[*Connection]bool)
	}
	manager.rooms[c.roomID][c] = true
}

func removeClientFromRoom(c *Connection) {
	manager.Lock()
	defer manager.Unlock()

	if clients, ok := manager.rooms[c.roomID]; ok {
		delete(clients, c)
		if len(clients) == 0 {
			delete(manager.rooms, c.roomID)
		}
	}
}

func broadcastToOthers(sender *Connection, msg Message) {
	manager.Lock()
	defer manager.Unlock()

	for client := range manager.rooms[sender.roomID] {
		if client == sender {
			continue
		}
		err := client.ws.WriteJSON(map[string]interface{}{
			"event": "changes",
			"data":  msg.Data, // Quill Delta JSON
		})
		if err != nil {
			log.Println("Write error:", err)
			client.ws.Close()
			delete(manager.rooms[sender.roomID], client)
		}
	}
}

func saveMessage(docId string, message json.RawMessage) {
	log.Printf("%v", message)
	documentID, err := uuid.Parse(docId)
	if err != nil {
		log.Printf("Invalid document ID: %s, error: %v", docId, err)
		return
	}

	// First, check if the document exists
	var doc models.Document
	if err := db.Where("id = ?", documentID).First(&doc).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Printf("Document %s not found", docId)
		} else {
			log.Printf("Database error when checking document %s: %v", docId, err)
		}
		return
	}

	// Update the document content
	if err := db.Model(&doc).Update("content", message).Error; err != nil {
		log.Printf("Failed to save document %s: %v", docId, err)
		return
	}

	log.Printf("Document %s saved successfully", docId)
}