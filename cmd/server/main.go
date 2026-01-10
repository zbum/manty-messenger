package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"

	"Mmessenger/internal/config"
	"Mmessenger/internal/database"
	"Mmessenger/internal/handler"
	"Mmessenger/internal/middleware"
	"Mmessenger/internal/repository"
	"Mmessenger/internal/service"
	"Mmessenger/internal/websocket"
	"Mmessenger/pkg/jwt"
)

func main() {
	// Load config
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Connect to database
	db, err := database.NewMySQL(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	log.Println("Connected to database")

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
	hub := websocket.NewHub()
	go hub.Run()

	// Initialize handlers
	authHandler := handler.NewAuthHandler(authService)
	roomHandler := handler.NewRoomHandler(roomService, hub)
	messageHandler := handler.NewMessageHandler(messageService)
	userHandler := handler.NewUserHandler(userRepo)

	// Initialize WebSocket handler
	wsHandler := websocket.NewHandler(hub, jwtService, messageService, memberRepo, userRepo, roomRepo)

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(jwtService)
	corsMiddleware := middleware.NewCORSMiddleware(cfg.CORS.AllowedOrigins)

	// Setup router
	r := mux.NewRouter()

	// Apply CORS middleware
	r.Use(corsMiddleware.Handler)

	// API routes
	api := r.PathPrefix("/api/v1").Subrouter()

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

	// WebSocket route
	r.HandleFunc("/ws", wsHandler.ServeWS)

	// Serve static files for frontend
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./frontend/dist")))

	// Start server
	addr := cfg.Server.Host + ":" + cfg.Server.Port
	log.Printf("Server starting on %s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
