package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"golang.org/x/crypto/bcrypt"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"
	_ "github.com/lib/pq"
)

// Client represents a connected WebSocket client.
type Client struct {
	UserID string
	Conn   *websocket.Conn
}

// Config contains database connection information.
type Config struct {
	DBUser     string `json:"db_user"`
	DBPassword string `json:"db_password"`
}

var (
	db       *sql.DB
	rdb      *redis.Client
	ctx      = context.Background()
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	clients   = make(map[string]*Client)
	broadcast = make(chan Message)
)

// Message represents a chat message.
type Message struct {
	ID        string `json:"id"`
	Sender    string `json:"sender"`
	Receiver  string `json:"receiver"`
	Content   string `json:"content"`
	Upvotes   int    `json:"upvotes"`
	Downvotes int    `json:"downvotes"`
}

func main() {
	var err error

	// Load the configuration file for Postgres username and password
	file, err := os.Open("config.json")
	if err != nil {
		log.Fatalf("Error opening config file: %v", err)
	}
	defer file.Close()

	configData, err := ioutil.ReadAll(file)
	if err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}

	var config Config
	if err := json.Unmarshal(configData, &config); err != nil {
		log.Fatalf("Error parsing config file: %v", err)
	}

	connStr := fmt.Sprintf("host=postgres user=%s password=%s sslmode=disable", config.DBUser, config.DBPassword)

	// Create the database named chat if it doesn't exist
	err = createDatabaseIfNotExists(connStr, "chat")
	if err != nil {
		log.Fatalf("Error creating database: %v", err)
	}
	fmt.Println("Database 'chat' created successfully")
	connStr += " dbname=chat"

	// Connect to the PostgreSQL database.
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatalf("Cannot connect to database: %v", err)
	}
	fmt.Println("Successfully connected to the database")

	// Run SQL migrations to create necessary tables.
	err = createTableUsers("create_table_users.sql")
	if err != nil {
		log.Fatalf("Error executing SQL migration for users: %v", err)
	}
	fmt.Println("Users table created successfully")

	err = createTableMessages("create_table_messages.sql")
	if err != nil {
		log.Fatalf("Error executing SQL migration for messages: %v", err)
	}
	fmt.Println("Messages table created successfully")

	err = createTableUserVotes("create_table_user_votes.sql")
	if err != nil {
		log.Fatalf("Error executing SQL migration for user_votes: %v", err)
	}
	fmt.Println("user_votes table created successfully")

	// Connect to Redis.
	rdb = redis.NewClient(&redis.Options{
		Addr: "redis:6379",
	})

	// Set up the Gin router with CORS.
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type"},
		AllowCredentials: true,
	}))

	// Defined the routes.
	r.POST("/signup", signupHandler)
	r.POST("/login", loginHandler)
	r.GET("/users", usersHandler)
	r.POST("/messages", sendMessageHandler)
	r.GET("/messages", getMessagesHandler)
	r.POST("/messages/:id/upvote", upvoteMessageHandler)
	r.POST("/messages/:id/downvote", downvoteMessageHandler)
	r.GET("/ws", wsHandler)

	// Start a goroutine to handle broadcasting messages to clients.
	go handleMessages()

	// Start the HTTP server.
	r.Run("0.0.0.0:8080")
}

// handleMessages broadcasts messages to the relevant clients.
func handleMessages() {
	for {
		msg := <-broadcast
		sendMessageToUser(msg.Sender, msg)
		sendMessageToUser(msg.Receiver, msg)
	}
}

// handleMessages broadcasts messages to the relevant clients.
func sendMessageToUser(userID string, msg Message) {
	client, exists := clients[userID]
	if exists {
		err := client.Conn.WriteJSON(msg)
		if err != nil {
			log.Printf("WebSocket error: %v", err)
			client.Conn.Close()
			delete(clients, userID)
		}
	}
}

// createDatabaseIfNotExists creates the specified database if it doesn't exist.
func createDatabaseIfNotExists(connStr, dbName string) error {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("error connecting to PostgreSQL: %v", err)
	}
	defer db.Close()

	var dbExists bool
	err = db.QueryRow("SELECT EXISTS (SELECT 1 FROM pg_database WHERE datname = $1)", dbName).Scan(&dbExists)
	if err != nil {
		return fmt.Errorf("error checking if database exists: %v", err)
	}

	if !dbExists {
		_, err = db.Exec(fmt.Sprintf("CREATE DATABASE %s", dbName))
		if err != nil {
			return fmt.Errorf("error creating database: %v", err)
		}
		fmt.Printf("Database '%s' created\n", dbName)
	}

	return nil
}

// createTableUsers creates the users table.
func createTableUsers(filepath string) error {
	return createTable(filepath, "users")
}

// createTableMessages creates the messages table.
func createTableMessages(filepath string) error {
	return createTable(filepath, "messages")
}

// createTableUserVotes creates the user_votes table.
func createTableUserVotes(filepath string) error {
	return createTable(filepath, "user_votes")
}

// createTable creates a table based on the provided SQL file.
func createTable(filepath, tableName string) error {
	var tableExists bool
	err := db.QueryRow(fmt.Sprintf("SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = '%s')", tableName)).Scan(&tableExists)
	if err != nil {
		return fmt.Errorf("error checking if table '%s' exists: %v", tableName, err)
	}

	if tableExists {
		fmt.Printf("Table '%s' already exists, skipping migration\n", tableName)
		return nil
	}

	sqlBytes, err := ioutil.ReadFile(filepath)
	if err != nil {
		return fmt.Errorf("failed to read SQL file: %v", err)
	}

	_, err = db.Exec(string(sqlBytes))
	if err != nil {
		return fmt.Errorf("failed to execute SQL statements: %v", err)
	}

	fmt.Printf("Migration executed successfully for table '%s'\n", tableName)

	return nil
}

// signupHandler handles user signup requests.
func signupHandler(c *gin.Context) {
	var user struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if user.Username == "" || user.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid username or password"})
		return
	}

	if len(user.Password) < 8 || len(user.Password) > 20 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Password must be between 8 to 20 characters."})
		return
	}

	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM users WHERE username = $1", user.Username).Scan(&count)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error checking username availability"})
		return
	}

	if count > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "Username already taken"})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	_, err = db.Exec("INSERT INTO users (username, password) VALUES ($1, $2)", user.Username, hashedPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User signed up successfully"})
}

// loginHandler handles user login requests.
func loginHandler(c *gin.Context) {
	var user struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	var storedPassword string
	err := db.QueryRow("SELECT password FROM users WHERE username = $1", user.Username).Scan(&storedPassword)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(user.Password))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Login successful"})
}

// usersHandler handles fetching all users.
func usersHandler(c *gin.Context) {
	currentUser := c.Query("username")

	rows, err := db.Query("SELECT username FROM users WHERE username != $1", currentUser)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}
	defer rows.Close()

	var users []string
	for rows.Next() {
		var username string
		if err := rows.Scan(&username); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan user"})
			return
		}
		users = append(users, username)
	}

	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error occurred during rows iteration"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"users": users})
}

// wsHandler handles WebSocket connections.
func wsHandler(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		http.NotFound(c.Writer, c.Request)
		return
	}
	defer conn.Close()

	userID := c.Query("user_id")
	client := &Client{UserID: userID, Conn: conn}
	clients[userID] = client

	for {
		var msg Message
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Printf("WebSocket read error: %v", err)
			delete(clients, userID)
			break
		}

		broadcast <- msg
	}
}

// sendMessageHandler handles sending messages.
func sendMessageHandler(c *gin.Context) {
	var msg Message

	if err := c.ShouldBindJSON(&msg); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var id int
	err := db.QueryRow(
		"INSERT INTO messages (sender, receiver, content, upvotes, downvotes) VALUES ($1, $2, $3, 0, 0) RETURNING id",
		msg.Sender, msg.Receiver, msg.Content,
	).Scan(&id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send message"})
		return
	}

	msg.ID = fmt.Sprintf("%d", id)
	broadcast <- msg

	c.JSON(http.StatusCreated, gin.H{"message": msg})
}

// getMessagesHandler handles fetching all messages.
func getMessagesHandler(c *gin.Context) {
	sender := c.Query("sender")
	receiver := c.Query("receiver")

	rows, err := db.Query(`
        SELECT id, sender, receiver, content, upvotes, downvotes
        FROM messages 
        WHERE (sender = $1 AND receiver = $2) OR (sender = $2 AND receiver = $1)
		ORDER BY timestamp
    `, sender, receiver)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch messages", "details": err.Error()})
		return
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var msg Message
		if err := rows.Scan(&msg.ID, &msg.Sender, &msg.Receiver, &msg.Content, &msg.Upvotes, &msg.Downvotes); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan message", "details": err.Error()})
			return
		}
		messages = append(messages, msg)
	}

	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error occurred during rows iteration", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"messages": messages})
}

// upvoteMessageHandler handles upvoting messages.
func upvoteMessageHandler(c *gin.Context) {
	userId := c.Query("user_id")
	messageId := c.Param("id")

	tx, err := db.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
		return
	}

	var existingVote string
	err = tx.QueryRow(`SELECT vote_type FROM user_votes WHERE user_id = $1 AND message_id = $2`, userId, messageId).Scan(&existingVote)
	if err != nil && err != sql.ErrNoRows {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch existing vote"})
		return
	}

	var upvotesBefore, downvotesBefore int

	err = tx.QueryRow(`SELECT upvotes, downvotes FROM messages WHERE id = $1`, messageId).Scan(&upvotesBefore, &downvotesBefore)
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch message votes"})
		return
	}

	var upvotesChange, downvotesChange int
	if existingVote == "upvote" {
		_, err = tx.Exec(`DELETE FROM user_votes WHERE user_id = $1 AND message_id = $2`, userId, messageId)
		upvotesChange = -1
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove upvote"})
			return
		}
	} else if existingVote == "downvote" {
		_, err = tx.Exec(`UPDATE user_votes SET vote_type = 'upvote' WHERE user_id = $1 AND message_id = $2`, userId, messageId)
		upvotesChange = 1
		downvotesChange = -1
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update vote"})
			return
		}
	} else {
		_, err = tx.Exec(`INSERT INTO user_votes (user_id, message_id, vote_type) VALUES ($1, $2, 'upvote')`, userId, messageId)
		upvotesChange = 1
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add upvote"})
			return
		}
	}

	_, err = tx.Exec(`UPDATE messages SET upvotes = upvotes + $1, downvotes = downvotes + $2 WHERE id = $3`, upvotesChange, downvotesChange, messageId)
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update vote counts"})
		return
	}

	rdb.HSet(ctx, fmt.Sprintf("message:%s", messageId), "upvotes", upvotesBefore+upvotesChange)
	rdb.HSet(ctx, fmt.Sprintf("message:%s", messageId), "downvotes", downvotesBefore+downvotesChange)

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}

	var updatedMessage Message
	err = db.QueryRow(`SELECT id, sender, receiver, content, upvotes, downvotes FROM messages WHERE id = $1`, messageId).Scan(&updatedMessage.ID, &updatedMessage.Sender, &updatedMessage.Receiver, &updatedMessage.Content, &updatedMessage.Upvotes, &updatedMessage.Downvotes)
	if err == nil {
		broadcast <- updatedMessage
	}

	c.JSON(http.StatusOK, gin.H{"message": "Vote toggled successfully"})
}

// downvoteMessageHandler handles downvoting messages.
func downvoteMessageHandler(c *gin.Context) {
	userId := c.Query("user_id")
	messageId := c.Param("id")

	tx, err := db.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
		return
	}

	var existingVote string
	err = tx.QueryRow(`SELECT vote_type FROM user_votes WHERE user_id = $1 AND message_id = $2`, userId, messageId).Scan(&existingVote)
	if err != nil && err != sql.ErrNoRows {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch existing vote"})
		return
	}

	var upvotesBefore, downvotesBefore int

	err = tx.QueryRow(`SELECT upvotes, downvotes FROM messages WHERE id = $1`, messageId).Scan(&upvotesBefore, &downvotesBefore)
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch message votes"})
		return
	}

	var upvotesChange, downvotesChange int
	if existingVote == "downvote" {
		_, err = tx.Exec(`DELETE FROM user_votes WHERE user_id = $1 AND message_id = $2`, userId, messageId)
		downvotesChange = -1
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove downvote"})
			return
		}
	} else if existingVote == "upvote" {
		_, err = tx.Exec(`UPDATE user_votes SET vote_type = 'downvote' WHERE user_id = $1 AND message_id = $2`, userId, messageId)
		upvotesChange = -1
		downvotesChange = 1
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update vote"})
			return
		}
	} else {
		_, err = tx.Exec(`INSERT INTO user_votes (user_id, message_id, vote_type) VALUES ($1, $2, 'downvote')`, userId, messageId)
		downvotesChange = 1
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add downvote"})
			return
		}
	}

	_, err = tx.Exec(`UPDATE messages SET upvotes = upvotes + $1, downvotes = downvotes + $2 WHERE id = $3`, upvotesChange, downvotesChange, messageId)
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update vote counts"})
		return
	}

	rdb.HSet(ctx, fmt.Sprintf("message:%s", messageId), "upvotes", upvotesBefore+upvotesChange)
	rdb.HSet(ctx, fmt.Sprintf("message:%s", messageId), "downvotes", downvotesBefore+downvotesChange)

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}

	var updatedMessage Message
	err = db.QueryRow(`SELECT id, sender, receiver, content, upvotes, downvotes FROM messages WHERE id = $1`, messageId).Scan(&updatedMessage.ID, &updatedMessage.Sender, &updatedMessage.Receiver, &updatedMessage.Content, &updatedMessage.Upvotes, &updatedMessage.Downvotes)
	if err == nil {
		broadcast <- updatedMessage
	}

	c.JSON(http.StatusOK, gin.H{"message": "Vote toggled successfully"})
}
