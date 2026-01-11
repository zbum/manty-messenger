# Mmessenger

WebSocket 기반 실시간 메신저 애플리케이션

## 기술 스택

- **Backend**: Go (gorilla/websocket, gorilla/mux)
- **Database**: MySQL
- **Frontend**: Vue.js 3 + Pinia + Vue Router
- **Authentication**: JWT (Access Token + Refresh Token)

## 주요 기능

- 사용자 인증 (회원가입, 로그인, 로그아웃)
- 채팅방 생성 및 관리
- 실시간 메시지 송수신
- 사용자 초대 (실시간 알림)
- 타이핑 표시
- 온라인 상태 표시

## 프로젝트 구조

```
Mmessenger/
├── cmd/server/main.go          # 서버 진입점
├── internal/
│   ├── config/                 # 환경설정
│   ├── database/               # DB 연결 및 마이그레이션
│   ├── models/                 # 데이터 모델
│   ├── repository/             # DB CRUD
│   ├── service/                # 비즈니스 로직
│   ├── handler/                # HTTP 핸들러
│   ├── websocket/              # WebSocket 처리
│   └── middleware/             # 인증, CORS 등
├── pkg/jwt/                    # JWT 유틸리티
└── frontend/                   # Vue.js 앱
```

## 설치 및 실행

### 사전 요구사항

- Go 1.21+
- Node.js 18+
- MySQL 8.0+

### 1. 데이터베이스 설정

```bash
mysql -h 127.0.0.1 -u root -p -e "CREATE DATABASE manty_messenger CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"
```

### 2. 환경 변수 설정

`.env` 파일 생성:

```env
DB_HOST=127.0.0.1
DB_PORT=3306
DB_USER=root
DB_PASS=your_password
DB_NAME=manty_messenger

JWT_SECRET=your-secret-key
JWT_ACCESS_EXPIRY=15m
JWT_REFRESH_EXPIRY=168h

SERVER_HOST=localhost
SERVER_PORT=8080

CORS_ORIGINS=http://localhost:5173
```

### 3. 데이터베이스 마이그레이션

```bash
mysql -h 127.0.0.1 -u root -p manty_messenger < internal/database/migrations/001_init.sql
```

### 4. 백엔드 실행

```bash
go run cmd/server/main.go
```

### 5. 프론트엔드 실행

```bash
cd frontend
npm install
npm run dev
```

## API 엔드포인트

### REST API

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/auth/register` | 회원가입 |
| POST | `/api/v1/auth/login` | 로그인 |
| POST | `/api/v1/auth/refresh` | 토큰 갱신 |
| POST | `/api/v1/auth/logout` | 로그아웃 |
| GET | `/api/v1/auth/me` | 내 정보 |
| GET | `/api/v1/rooms` | 채팅방 목록 |
| POST | `/api/v1/rooms` | 채팅방 생성 |
| GET | `/api/v1/rooms/:id/messages` | 메시지 조회 |
| POST | `/api/v1/rooms/:id/members` | 멤버 초대 |

### WebSocket

연결: `ws://localhost:8080/ws?token=<jwt>`

| Type | Direction | Description |
|------|-----------|-------------|
| `join_room` | Client → Server | 채팅방 입장 |
| `leave_room` | Client → Server | 채팅방 퇴장 |
| `send_message` | Client → Server | 메시지 전송 |
| `typing` | Client → Server | 타이핑 상태 |
| `new_message` | Server → Client | 새 메시지 수신 |
| `user_joined` | Server → Client | 사용자 입장 알림 |
| `user_left` | Server → Client | 사용자 퇴장 알림 |
| `room_invited` | Server → Client | 채팅방 초대 알림 |

## 라이선스

MIT
