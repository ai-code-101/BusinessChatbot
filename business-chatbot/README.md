# Business AI Chatbot - Backend

A multi-tenant business chatbot backend. Admins upload documents or paste
text to build their business's knowledge base; customers ask questions
through `/v1/chat/ask` and get answers grounded in that business's content
only (RAG - Retrieval Augmented Generation).

## Architecture

```
Frontend  --->  Go/Gin (backend/)  --->  Python/FastAPI (rag-service/)
                     |                          |
                  MongoDB                   ChromaDB (vectors)
                (metadata)                       |
                                          GitHub Models (LLM)
```

- **backend/** - Go/Gin API. Handles auth (JWT), admin document management,
  and the public chat endpoint. Talks to MongoDB for metadata and to the
  RAG service internally for anything AI-related. No AI logic lives here.
- **rag-service/** - Python/FastAPI. Handles chunking, embeddings (local
  sentence-transformers model), vector storage (ChromaDB, one collection
  per business), and calls GitHub Models to generate answers.

## Running locally with Docker Compose

1. Copy env files and fill in real values:
   ```bash
   cp backend/.env.example backend/.env
   cp rag-service/.env.example rag-service/.env
   ```
2. Add your GitHub personal access token (with `models: read` permission)
   to `rag-service/.env` as `GITHUB_TOKEN`.
3. Start everything:
   ```bash
   docker compose up --build
   ```
4. Services will be available at:
   - Backend API: `http://localhost:8080/v1`
   - RAG service (internal only, but reachable for debugging): `http://localhost:8000`
   - MongoDB: `localhost:27017`

## Trying it out with curl

**Admin login** (uses the seeded admin from `.env`):
```bash
curl -X POST http://localhost:8080/v1/admin/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"change_this_password"}'
```

**Paste text as a knowledge source** (use the token from login):
```bash
curl -X POST http://localhost:8080/v1/admin/documents/text \
  -H "Authorization: Bearer <TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{"title":"Delivery Policy","content":"We deliver within Nairobi in 2-3 business days..."}'
```

**Ask the chatbot a question** (public endpoint, no auth needed):
```bash
curl -X POST http://localhost:8080/v1/chat/ask \
  -H "Content-Type: application/json" \
  -d '{"business_id":"default_business","question":"How long does delivery take?"}'
```

## API summary

| Method | Endpoint | Auth | Purpose |
|---|---|---|---|
| POST | `/v1/admin/login` | none | Admin login, returns JWT |
| POST | `/v1/admin/documents/upload` | admin JWT | Upload a text file |
| POST | `/v1/admin/documents/text` | admin JWT | Paste text directly |
| GET | `/v1/admin/documents` | admin JWT | List this business's documents |
| DELETE | `/v1/admin/documents/:id` | admin JWT | Delete a document |
| POST | `/v1/chat/ask` | none | Ask the chatbot a question |
