package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"os"
	"net/http"
	"time"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoDB конфигурация
const (
	// mongoURI = "mongodb+srv://nagidevfullstack:Hadi2017g@golangdb.ebvlakf.mongodb.net/?retryWrites=true&w=majority&appName=GolangDB"
	mongoURI     = os.Getenv("mongodb+srv://nagidevfullstack:Hadi2017g@golangdb.ebvlakf.mongodb.net/?retryWrites=true&w=majority&appName=GolangDB")
	databaseName = "TodoListONGolang"
	collectionUser    = "users"
	collectionPost    = "posts"
	collectionChat    = "chats"
	collectionMessage = "messages"
	collectionNotice  = "notices"
)

var client *mongo.Client
var cld *cloudinary.Cloudinary

// Структуры, эквивалентные схемам Mongoose
type User struct {
	ID               primitive.ObjectID `json:"_id" bson:"_id"`
	GoogleID         string             `json:"googleId" bson:"googleId"`
	Name             string             `json:"name" bson:"name"`
	Email            string             `json:"email" bson:"email"`
	Avatar           string             `json:"avatar" bson:"avatar"`
	RegistrationDate string             `json:"registrationDate" bson:"registrationDate"`
	Subscriptions    []Subscription     `json:"subscriptions" bson:"subscriptions"`
	Subscribers      []Subscriber       `json:"subscribers" bson:"subscribers"`
	LikesPosts       []LikePost         `json:"likesPosts" bson:"likesPosts"`
	Messages         []UserMessage      `json:"messages" bson:"messages"`
	Bookmarks        []Bookmark         `json:"bookmarks" bson:"bookmarks"`
	Reposts          []Repost           `json:"reposts" bson:"reposts"`
	Posts            []UserPost         `json:"posts" bson:"posts"`
}

type Subscription struct {
	User   string `json:"user" bson:"user"`
	Avatar string `json:"avatar" bson:"avatar"`
	Name   string `json:"name" bson:"name"`
}

type Subscriber struct {
	User   string `json:"user" bson:"user"`
	Avatar string `json:"avatar" bson:"avatar"`
	Name   string `json:"name" bson:"name"`
}

type LikePost struct {
	Post   string `json:"post" bson:"post"`
	Author string `json:"author" bson:"author"`
	State  bool   `json:"state" bson:"state"`
}

type UserMessage struct {
	Author     string `json:"author" bson:"author"`
	MessagesID string `json:"messagesID" bson:"messagesID"`
}

type Bookmark struct {
	Post   string `json:"post" bson:"post"`
	Author string `json:"author" bson:"author"`
	State  bool   `json:"state" bson:"state"`
}

type Repost struct {
	Post   string `json:"post" bson:"post"`
	Author string `json:"author" bson:"author"`
	State  bool   `json:"state" bson:"state"`
}

type UserPost struct {
	Post   string `json:"post" bson:"post"`
	Author string `json:"author" bson:"author"`
}

type Post struct {
	ID         primitive.ObjectID `json:"_id" bson:"_id"`
	Author     string             `json:"author" bson:"author"`
	Text       string             `json:"text" bson:"text"`
	Images     string             `json:"images" bson:"images"`
	CreateDate string             `json:"createDate" bson:"createDate"`
	Likes      int                `json:"likes" bson:"likes"`
	Comments   []Comment          `json:"comments" bson:"comments"`
	Reposts    []PostRepost       `json:"reposts" bson:"reposts"`
	Bookmarks  []PostBookmark     `json:"bookmarks" bson:"bookmarks"`
}

type Comment struct {
	Text       string    `json:"text" bson:"text"`
	Author     string    `json:"author" bson:"author"`
	CreateDate time.Time `json:"createDate" bson:"createDate"`
}

type PostRepost struct {
	PostID     string `json:"post_id" bson:"post_id"`
	Author     string `json:"author" bson:"author"`
	CreateDate string `json:"createDate" bson:"createDate"`
}

type PostBookmark struct {
	PostID     string    `json:"post_id" bson:"post_id"`
	Author     string    `json:"author" bson:"author"`
	CreateDate time.Time `json:"createDate" bson:"createDate"`
}

type Chat struct {
	ID         primitive.ObjectID `json:"_id" bson:"_id"`
	IDD        string             `json:"idd" bson:"idd"`
	Text       string             `json:"text" bson:"text"`
	Author     string             `json:"author" bson:"author"`
	CreateDate string             `json:"createDate" bson:"createDate"`
	Img        string             `json:"img" bson:"img"`
}

type Message struct {
	ID         primitive.ObjectID `json:"_id" bson:"_id"`
	IDField    string             `json:"id" bson:"id"`
	Sender     string             `json:"sender" bson:"sender"`
	Receiver   string             `json:"receiver" bson:"receiver"`
	CreateDate string             `json:"createDate" bson:"createDate"`
	Img        string             `json:"img" bson:"img"` // Добавлено поле Img
}

type Notice struct {
	ID         primitive.ObjectID `json:"_id" bson:"_id"`
	User       string             `json:"user" bson:"user"`
	Type       string             `json:"type" bson:"type"`
	Post       string             `json:"post" bson:"post"`
	FromUser   []FromUser         `json:"fromUser" bson:"fromUser"`
	CreateDate time.Time          `json:"createDate" bson:"createDate"`
	Read       bool               `json:"read" bson:"read"`
}

type FromUser struct {
	ID     string `json:"id_" bson:"id_"`
	IDUser string `json:"id_user" bson:"id_user"`
}

// Middleware для CORS
func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Accept")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func main() {
	// Подключение к MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var err error
	client, err = mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Connected to MongoDB!")

	// Настройка Cloudinary
	// cld, err = cloudinary.NewFromParams("ddtq1ack5", "845634458425448", "ZTt9tU5JtlAhH5pwfYIU7dMYmzU")
	cld, err = cloudinary.NewFromParams(
		os.Getenv("ddtq1ack5"),
		os.Getenv("845634458425448"),
		os.Getenv("ZTt9tU5JtlAhH5pwfYIU7dMYmzU"),
	)
	if err != nil {
		log.Fatal(err)
	}

	// Создание маршрутизатора
	router := mux.NewRouter()

	// Базовый маршрут
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello from API")
	}).Methods("GET")

	// Маршруты API
	api := router.PathPrefix("/api/twitter").Subrouter()

	// User Routes
	api.HandleFunc("/users", getUsers).Methods("GET", "OPTIONS")
	api.HandleFunc("/users", createUser).Methods("POST", "OPTIONS")
	api.HandleFunc("/users/check-existence", checkUserExistence).Methods("POST", "OPTIONS")
	api.HandleFunc("/users/{id}", deleteUser).Methods("DELETE", "OPTIONS")
	api.HandleFunc("/users/{id}", updateUser).Methods("PUT", "OPTIONS")
	api.HandleFunc("/users/{googleId}", getUserByGoogleID).Methods("GET", "OPTIONS")

	// Post Routes
	api.HandleFunc("/posts", getPosts).Methods("GET", "OPTIONS")
	api.HandleFunc("/posts", createPost).Methods("POST", "OPTIONS")
	api.HandleFunc("/posts/{id}", getPostByID).Methods("GET", "OPTIONS")
	api.HandleFunc("/posts/{id}", deletePost).Methods("DELETE", "OPTIONS")
	api.HandleFunc("/posts/{id}", updatePost).Methods("PUT", "OPTIONS")

	// Chat Routes
	api.HandleFunc("/chat", getChats).Methods("GET", "OPTIONS")
	api.HandleFunc("/chat", createChat).Methods("POST", "OPTIONS")
	api.HandleFunc("/chat/{idd}", getChatsByIDD).Methods("GET", "OPTIONS")

	// Message Routes
	api.HandleFunc("/messages", getMessages).Methods("GET", "OPTIONS")
	api.HandleFunc("/messages", sendMessage).Methods("POST", "OPTIONS")
	api.HandleFunc("/messages/{id}", getMessageByID).Methods("GET", "OPTIONS")
	api.HandleFunc("/messages/{id}", deleteMessage).Methods("DELETE", "OPTIONS")
	api.HandleFunc("/messages/{id}", updateMessage).Methods("PUT", "OPTIONS")

	// Notice Routes
	api.HandleFunc("/notices", getNotices).Methods("GET", "OPTIONS")
	api.HandleFunc("/notices", createNotice).Methods("POST", "OPTIONS")
	api.HandleFunc("/notices/{id}", deleteNotice).Methods("DELETE", "OPTIONS")
	api.HandleFunc("/notices/{id}", updateNotice).Methods("PUT", "OPTIONS")

	// Добавляем middleware для CORS
	corsRouter := enableCORS(router)

	// Запуск сервера
	fmt.Println("Server is running on port 7070...")
	// log.Fatal(http.ListenAndServe(":7070", corsRouter))
	port := os.Getenv("PORT")
if port == "" {
    port = "7070" // Локальный порт по умолчанию
}
log.Fatal(http.ListenAndServe("0.0.0.0:"+port, corsRouter))
}

// --- User Handlers ---

func createUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if user.Email == "" || user.GoogleID == "" || user.Name == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	collection := client.Database(databaseName).Collection(collectionUser)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Проверка существующего пользователя
	var existingUser User
	err := collection.FindOne(ctx, bson.M{"email": user.Email}).Decode(&existingUser)
	if err == nil {
		http.Error(w, "User already exists", http.StatusBadRequest)
		return
	}

	user.ID = primitive.NewObjectID()
	if user.RegistrationDate == "" {
		user.RegistrationDate = time.Now().Format(time.RFC3339)
	}
	user.Subscriptions = []Subscription{}
	user.Subscribers = []Subscriber{}
	user.LikesPosts = []LikePost{}
	user.Messages = []UserMessage{}
	user.Bookmarks = []Bookmark{}
	user.Reposts = []Repost{}
	user.Posts = []UserPost{}

	_, err = collection.InsertOne(ctx, user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

func getUsers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	collection := client.Database(databaseName).Collection(collectionUser)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)

	var users []User
	if err = cursor.All(ctx, &users); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(users)
}

func checkUserExistence(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var body struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	collection := client.Database(databaseName).Collection(collectionUser)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var user User
	err := collection.FindOne(ctx, bson.M{"email": body.Email}).Decode(&user)
	exists := err == nil

	json.NewEncoder(w).Encode(map[string]bool{"exists": exists})
}

func getUserByGoogleID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	params := mux.Vars(r)
	googleID := params["googleId"]

	collection := client.Database(databaseName).Collection(collectionUser)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var user User
	err := collection.FindOne(ctx, bson.M{"googleId": googleID}).Decode(&user)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(user)
}

func deleteUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	params := mux.Vars(r)
	id, err := primitive.ObjectIDFromHex(params["id"])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	collection := client.Database(databaseName).Collection(collectionUser)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if result.DeletedCount == 0 {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"message": "User deleted successfully"})
}

func updateUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	params := mux.Vars(r)
	id, err := primitive.ObjectIDFromHex(params["id"])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var updates User
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	collection := client.Database(databaseName).Collection(collectionUser)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	update := bson.M{"$set": updates}
	var updatedUser User
	err = collection.FindOneAndUpdate(ctx, bson.M{"_id": id}, update, options.FindOneAndUpdate().SetReturnDocument(options.After)).Decode(&updatedUser)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(updatedUser)
}

// --- Post Handlers ---
func createPost(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    var post Post
    contentType := r.Header.Get("Content-Type")

    // Обработка multipart/form-data
    if strings.HasPrefix(contentType, "multipart/form-data") {
        err := r.ParseMultipartForm(10 << 20) // 10 MB limit
        if err != nil {
            log.Printf("Error parsing form: %v", err)
            http.Error(w, "Unable to parse form", http.StatusBadRequest)
            return
        }

        post.Author = r.FormValue("author")
        post.Text = r.FormValue("text")

        // Проверка обязательных полей
        if post.Author == "" || post.Text == "" {
            log.Println("Missing author or text")
            http.Error(w, "Author and text are required", http.StatusBadRequest)
            return
        }

        // Обработка изображения
        file, _, err := r.FormFile("images")
        if err == nil {
            defer file.Close()
            uploadResult, err := cld.Upload.Upload(ctx, file, uploader.UploadParams{
                Folder:         "social-network",
                AllowedFormats: []string{"jpg", "png", "webp", "gif"},
                Transformation: "w_1200,c_limit",
            })
            if err != nil {
                log.Printf("Error uploading image: %v", err)
                http.Error(w, "Failed to upload image", http.StatusInternalServerError)
                return
            }
            post.Images = uploadResult.SecureURL
        }
    } else {
        // Обработка JSON
        if err := json.NewDecoder(r.Body).Decode(&post); err != nil {
            log.Printf("Error decoding JSON: %v", err)
            http.Error(w, "Invalid request body", http.StatusBadRequest)
            return
        }

        // Проверка обязательных полей
        if post.Author == "" || post.Text == "" {
            log.Println("Missing author or text")
            http.Error(w, "Author and text are required", http.StatusBadRequest)
            return
        }
    }

    // Инициализация полей поста
    post.ID = primitive.NewObjectID()
    post.Likes = 0
    post.Comments = []Comment{}
    post.Reposts = []PostRepost{}
    post.Bookmarks = []PostBookmark{}
    if post.CreateDate == "" {
        post.CreateDate = time.Now().Format(time.RFC3339)
    }

    // Сохранение в MongoDB
    collection := client.Database(databaseName).Collection(collectionPost)
    _, err := collection.InsertOne(ctx, post)
    if err != nil {
        log.Printf("Error inserting post: %v", err)
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(post)
}

func getPosts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	collection := client.Database(databaseName).Collection(collectionPost)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)

	var posts []Post
	if err = cursor.All(ctx, &posts); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(posts)
}

func getPostByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	params := mux.Vars(r)
	id, err := primitive.ObjectIDFromHex(params["id"])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	collection := client.Database(databaseName).Collection(collectionPost)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var post Post
	err = collection.FindOne(ctx, bson.M{"_id": id}).Decode(&post)
	if err != nil {
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(post)
}

func deletePost(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	params := mux.Vars(r)
	id, err := primitive.ObjectIDFromHex(params["id"])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	collection := client.Database(databaseName).Collection(collectionPost)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if result.DeletedCount == 0 {
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"message": "Post deleted successfully"})
}

func updatePost(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	params := mux.Vars(r)
	id, err := primitive.ObjectIDFromHex(params["id"])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var updates Post
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Создаём контекст для Cloudinary
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Обработка загрузки изображения
	err = r.ParseMultipartForm(10 << 20)
	if err == nil {
		file, _, err := r.FormFile("images")
		if err == nil {
			defer file.Close()
			uploadResult, err := cld.Upload.Upload(ctx, file, uploader.UploadParams{
				Folder:         "social-network",
				AllowedFormats: []string{"jpg", "png", "webp", "gif"},
				Transformation: "w_1200,c_limit",
			})
			if err != nil {
				http.Error(w, "Failed to upload image", http.StatusInternalServerError)
				return
			}
			updates.Images = uploadResult.SecureURL
		}
	}

	collection := client.Database(databaseName).Collection(collectionPost)
	update := bson.M{"$set": updates}
	var updatedPost Post
	err = collection.FindOneAndUpdate(ctx, bson.M{"_id": id}, update, options.FindOneAndUpdate().SetReturnDocument(options.After)).Decode(&updatedPost)
	if err != nil {
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(updatedPost)
}

// --- Chat Handlers ---

func createChat(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var chat Chat
	if err := json.NewDecoder(r.Body).Decode(&chat); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if chat.Author == "" || chat.Text == "" {
		http.Error(w, "Author and text are required", http.StatusBadRequest)
		return
	}

	// Создаём контекст для Cloudinary
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Обработка загрузки изображения
	err := r.ParseMultipartForm(10 << 20)
	if err == nil {
		file, _, err := r.FormFile("img")
		if err == nil {
			defer file.Close()
			uploadResult, err := cld.Upload.Upload(ctx, file, uploader.UploadParams{
				Folder:         "social-network",
				AllowedFormats: []string{"jpg", "png", "webp", "gif"},
				Transformation: "w_1200,c_limit",
			})
			if err != nil {
				http.Error(w, "Failed to upload image", http.StatusInternalServerError)
				return
			}
			chat.Img = uploadResult.SecureURL
		}
	}

	chat.ID = primitive.NewObjectID()

	collection := client.Database(databaseName).Collection(collectionChat)
	_, err = collection.InsertOne(ctx, chat)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(chat)
}

func getChats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	collection := client.Database(databaseName).Collection(collectionChat)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)

	var chats []Chat
	if err = cursor.All(ctx, &chats); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(chats)
}

func getChatsByIDD(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	params := mux.Vars(r)
	idd := params["idd"]

	collection := client.Database(databaseName).Collection(collectionChat)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := collection.Find(ctx, bson.M{"idd": idd})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)

	var chats []Chat
	if err = cursor.All(ctx, &chats); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(chats) == 0 {
		http.Error(w, "No chats found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(chats)
}

// --- Message Handlers ---

func sendMessage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var message Message
	if err := json.NewDecoder(r.Body).Decode(&message); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	message.ID = primitive.NewObjectID()

	collection := client.Database(databaseName).Collection(collectionMessage)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := collection.InsertOne(ctx, message)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(message)
}

func getMessages(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	collection := client.Database(databaseName).Collection(collectionMessage)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)

	var messages []Message
	if err = cursor.All(ctx, &messages); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(messages)
}

func getMessageByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	params := mux.Vars(r)
	id := params["id"]

	collection := client.Database(databaseName).Collection(collectionMessage)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var message Message
	err := collection.FindOne(ctx, bson.M{"id": id}).Decode(&message)
	if err != nil {
		http.Error(w, "Message not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(message)
}

func deleteMessage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	params := mux.Vars(r)
	id, err := primitive.ObjectIDFromHex(params["id"])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	collection := client.Database(databaseName).Collection(collectionMessage)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if result.DeletedCount == 0 {
		http.Error(w, "Message not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"message": "Message deleted successfully"})
}

func updateMessage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	params := mux.Vars(r)
	id, err := primitive.ObjectIDFromHex(params["id"])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var updates Message
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Создаём контекст для Cloudinary
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Обработка загрузки изображения
	err = r.ParseMultipartForm(10 << 20)
	if err == nil {
		file, _, err := r.FormFile("img")
		if err == nil {
			defer file.Close()
			uploadResult, err := cld.Upload.Upload(ctx, file, uploader.UploadParams{
				Folder:         "social-network",
				AllowedFormats: []string{"jpg", "png", "webp", "gif"},
				Transformation: "w_1200,c_limit",
			})
			if err != nil {
				http.Error(w, "Failed to upload image", http.StatusInternalServerError)
				return
			}
			updates.Img = uploadResult.SecureURL
		}
	}

	collection := client.Database(databaseName).Collection(collectionMessage)
	update := bson.M{"$set": updates}
	var updatedMessage Message
	err = collection.FindOneAndUpdate(ctx, bson.M{"_id": id}, update, options.FindOneAndUpdate().SetReturnDocument(options.After)).Decode(&updatedMessage)
	if err != nil {
		http.Error(w, "Message not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(updatedMessage)
}

// --- Notice Handlers ---

func createNotice(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var notice Notice
	if err := json.NewDecoder(r.Body).Decode(&notice); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	notice.ID = primitive.NewObjectID()
	notice.CreateDate = time.Now()
	notice.Read = false

	collection := client.Database(databaseName).Collection(collectionNotice)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := collection.InsertOne(ctx, notice)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(notice)
}

func getNotices(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	collection := client.Database(databaseName).Collection(collectionNotice)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)

	var notices []Notice
	if err = cursor.All(ctx, &notices); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(notices)
}

func deleteNotice(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	params := mux.Vars(r)
	id, err := primitive.ObjectIDFromHex(params["id"])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	collection := client.Database(databaseName).Collection(collectionNotice)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if result.DeletedCount == 0 {
		http.Error(w, "Notice not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"message": "Notice deleted successfully"})
}

func updateNotice(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	params := mux.Vars(r)
	id, err := primitive.ObjectIDFromHex(params["id"])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var updates Notice
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	collection := client.Database(databaseName).Collection(collectionNotice)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	update := bson.M{"$set": updates}
	var updatedNotice Notice
	err = collection.FindOneAndUpdate(ctx, bson.M{"_id": id}, update, options.FindOneAndUpdate().SetReturnDocument(options.After)).Decode(&updatedNotice)
	if err != nil {
		http.Error(w, "Notice not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(updatedNotice)
}