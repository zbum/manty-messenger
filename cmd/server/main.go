package main

import (
	"context"
	"log"
	"net/http"

	"github.com/gorilla/mux"

	"Mmessenger/internal/config"
	"Mmessenger/internal/database"
	"Mmessenger/internal/handler"
	"Mmessenger/internal/middleware"
	"Mmessenger/internal/pubsub"
	"Mmessenger/internal/repository"
	"Mmessenger/internal/service"
	"Mmessenger/internal/storage"
	"Mmessenger/internal/websocket"
	"Mmessenger/pkg/jwt"
)

func main() {
	log.Println("=== Mmessenger Server Starting ===")

	// Load config
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Print configuration
	log.Println("Configuration loaded:")
	log.Printf("  Server: %s:%s", cfg.Server.Host, cfg.Server.Port)
	log.Printf("  Database: %s@%s:%s/%s", cfg.Database.User, cfg.Database.Host, cfg.Database.Port, cfg.Database.Name)
	log.Printf("  CORS Origins: %v", cfg.CORS.AllowedOrigins)
	log.Printf("  JWT Access Expiry: %s", cfg.JWT.AccessExpiry)
	log.Printf("  JWT Refresh Expiry: %s", cfg.JWT.RefreshExpiry)

	// Connect to database
	log.Printf("Connecting to database %s:%s...", cfg.Database.Host, cfg.Database.Port)
	db, err := database.NewMySQL(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	log.Println("Database connection established")

	// Connect to Redis
	log.Printf("Connecting to Redis %s:%s...", cfg.Redis.Host, cfg.Redis.Port)
	redisClient, err := database.NewRedis(&cfg.Redis)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redisClient.Close()

	log.Println("Redis connection established")

	// Initialize Redis Pub/Sub
	redisPubSub := pubsub.NewRedisPubSub(redisClient)
	if err := redisPubSub.Subscribe(context.Background(),
		pubsub.ChannelRoomMessage,
		pubsub.ChannelUserMessage,
		pubsub.ChannelPresence,
	); err != nil {
		log.Fatalf("Failed to subscribe to Redis channels: %v", err)
	}
	defer redisPubSub.Close()

	log.Println("Redis Pub/Sub initialized")

	// Initialize file storage
	log.Printf("Initializing file storage at %s...", cfg.Storage.BasePath)
	localStorage, err := storage.NewLocalStorage(
		cfg.Storage.BasePath,
		cfg.Storage.BaseURL,
		cfg.Storage.MaxFileSize,
	)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}
	log.Println("File storage initialized")

	// Initialize thumbnail generator (libvips)
	storage.InitThumbnail()
	defer storage.ShutdownThumbnail()
	log.Println("Thumbnail generator initialized")

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	roomRepo := repository.NewRoomRepository(db)
	messageRepo := repository.NewMessageRepository(db)
	memberRepo := repository.NewRoomMemberRepository(db)

	// Initialize JWT service
	jwtService := jwt.NewService(&cfg.JWT)

	// Initialize services
	authService := service.NewAuthService(userRepo, jwtService, db)
	roomService := service.NewRoomService(roomRepo, memberRepo, userRepo)
	messageService := service.NewMessageService(messageRepo, memberRepo, userRepo)

	// Initialize WebSocket Hub first (needed by RoomHandler)
	hub := websocket.NewHub(redisPubSub)
	go hub.Run()

	// Initialize handlers
	authHandler := handler.NewAuthHandler(authService)
	roomHandler := handler.NewRoomHandler(roomService, hub)
	messageHandler := handler.NewMessageHandler(messageService)
	userHandler := handler.NewUserHandler(userRepo)
	fileHandler := handler.NewFileHandler(localStorage, cfg.Storage.MaxFileSize)

	// Initialize WebSocket handler
	wsHandler := websocket.NewHandler(hub, jwtService, messageService, memberRepo, userRepo, roomRepo)

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(jwtService)
	corsMiddleware := middleware.NewCORSMiddleware(cfg.CORS.AllowedOrigins)

	// Setup router
	r := mux.NewRouter()

	// Apply middleware
	r.Use(middleware.AccessLog)
	r.Use(corsMiddleware.Handler)

	// API routes
	api := r.PathPrefix("/api/v1").Subrouter()

	// Health check (public)
	api.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}).Methods("GET")

	// Auth routes (public)
	api.HandleFunc("/auth/register", authHandler.Register).Methods("POST")
	api.HandleFunc("/auth/login", authHandler.Login).Methods("POST")
	api.HandleFunc("/auth/refresh", authHandler.RefreshToken).Methods("POST")

	// Protected auth routes
	authProtected := api.PathPrefix("/auth").Subrouter()
	authProtected.Use(authMiddleware.Authenticate)
	authProtected.HandleFunc("/logout", authHandler.Logout).Methods("POST")
	authProtected.HandleFunc("/me", authHandler.Me).Methods("GET")

	// User routes (protected)
	userRoutes := api.PathPrefix("/users").Subrouter()
	userRoutes.Use(authMiddleware.Authenticate)
	userRoutes.HandleFunc("", userHandler.Search).Methods("GET")
	userRoutes.HandleFunc("/{id:[0-9]+}", userHandler.GetByID).Methods("GET")

	// Room routes (protected)
	roomRoutes := api.PathPrefix("/rooms").Subrouter()
	roomRoutes.Use(authMiddleware.Authenticate)
	roomRoutes.HandleFunc("", roomHandler.GetMyRooms).Methods("GET")
	roomRoutes.HandleFunc("", roomHandler.Create).Methods("POST")
	roomRoutes.HandleFunc("/{id:[0-9]+}", roomHandler.GetByID).Methods("GET")
	roomRoutes.HandleFunc("/{id:[0-9]+}", roomHandler.Update).Methods("PUT")
	roomRoutes.HandleFunc("/{id:[0-9]+}", roomHandler.Delete).Methods("DELETE")
	roomRoutes.HandleFunc("/{id:[0-9]+}/members", roomHandler.GetMembers).Methods("GET")
	roomRoutes.HandleFunc("/{id:[0-9]+}/members", roomHandler.AddMember).Methods("POST")
	roomRoutes.HandleFunc("/{id:[0-9]+}/members/{userId:[0-9]+}", roomHandler.RemoveMember).Methods("DELETE")
	roomRoutes.HandleFunc("/{id:[0-9]+}/leave", roomHandler.Leave).Methods("POST")

	// Message routes (protected)
	roomRoutes.HandleFunc("/{id:[0-9]+}/messages", messageHandler.GetMessages).Methods("GET")
	roomRoutes.HandleFunc("/{id:[0-9]+}/messages/{msgId:[0-9]+}", messageHandler.GetMessage).Methods("GET")
	roomRoutes.HandleFunc("/{id:[0-9]+}/messages/{msgId:[0-9]+}", messageHandler.Update).Methods("PUT")
	roomRoutes.HandleFunc("/{id:[0-9]+}/messages/{msgId:[0-9]+}", messageHandler.Delete).Methods("DELETE")

	// File routes (protected)
	fileRoutes := api.PathPrefix("/files").Subrouter()
	fileRoutes.Use(authMiddleware.Authenticate)
	fileRoutes.HandleFunc("/upload", fileHandler.Upload).Methods("POST")

	// WebSocket route
	r.HandleFunc("/ws", wsHandler.ServeWS)

	// Serve uploaded files
	r.PathPrefix(cfg.Storage.BaseURL + "/").Handler(
		http.StripPrefix(cfg.Storage.BaseURL+"/",
			http.FileServer(http.Dir(cfg.Storage.BasePath))))

	// Serve static files for frontend
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./frontend/dist")))

	// Start server
	addr := cfg.Server.Host + ":" + cfg.Server.Port
	log.Println("=== Server Initialization Complete ===")
	log.Printf("Listening on http://%s", addr)
	log.Println("Health check: http://" + addr + "/api/v1/health")
	log.Println("Ready to accept connections")

	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
