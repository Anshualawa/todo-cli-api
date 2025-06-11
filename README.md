# 📝 Go TODO REST API

A simple RESTful TODO API built in **Go** that uses a local `tasks.json` file as a lightweight data store. Supports creating and listing tasks with CORS-enabled middleware.

---

## 📦 Features

- ✅ Add new tasks via `POST /tasks`
- ✅ Get all tasks via `GET /tasks`
- ✅ Save/load tasks from `tasks.json`
- ✅ JSON response format
- ✅ CORS & Content-Type middleware
- 🚧 Ready for expansion: PUT, DELETE, CLI client, etc.

---

## 🚀 Getting Started

### 1. Clone the repository

```bash
git clone https://github.com/Anshualawa/todo-cli-api.git
cd todo-cli-api
```

## 📁 Project Structure
```bash
.
├── main.go         # Main server code
├── tasks.json      # Data file (auto-created)
```

## 📌 API Endpoints
### ✅ GET /tasks
- Get the list of all TODOs.

- Request:
http://localhost:8080/tasks

- Response :
```json
[
  {
    "id": 1,
    "title": "Learn Go",
    "completed": false
  },
  {
    "id": 2,
    "title": "Build a REST API",
    "completed": true
  }
]
```
### ✅ POST /tasks
- Create a new task.

Request:
```bash
curl -X POST http://localhost:8080/tasks
-H "Content-Type: application/json" -d '{"title":"Buy groceries","completed": false}'
```
Response :
```json
{
  "id": 3,
  "title": "Buy groceries",
  "completed": false
}
```
