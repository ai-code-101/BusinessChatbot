"""
Internal RAG microservice. This is never called directly by any frontend -
only by the Go backend. It handles:
  - /ingest   -> chunk + embed + store a document's text
  - /query    -> embed a question, retrieve relevant chunks, ask the LLM
  - /documents/{doc_id} (DELETE) -> remove a document's vectors
"""

from fastapi import FastAPI, HTTPException
from pydantic import BaseModel

from services.chunking import chunk_text
from services.embeddings import embed_texts
from services.vector_store import add_chunks, query as vector_query, delete_by_doc_id
from services.llm import generate_answer

app = FastAPI(title="Business RAG Service")


class IngestRequest(BaseModel):
    business_id: str
    doc_id: str
    title: str
    content: str


class IngestResponse(BaseModel):
    chunk_count: int


class QueryRequest(BaseModel):
    business_id: str
    question: str
    session_id: str = ""


class QueryResponse(BaseModel):
    answer: str
    sources: list[str]
    session_id: str
    tokens_used: int


@app.get("/health")
def health():
    return {"status": "ok"}


@app.post("/ingest", response_model=IngestResponse)
def ingest(req: IngestRequest):
    chunks = chunk_text(req.content)
    if not chunks:
        raise HTTPException(status_code=400, detail="content produced no chunks (empty or too short)")

    embeddings = embed_texts(chunks)
    add_chunks(req.business_id, req.doc_id, req.title, chunks, embeddings)

    return IngestResponse(chunk_count=len(chunks))


@app.post("/query", response_model=QueryResponse)
def query(req: QueryRequest):
    if not req.question.strip():
        raise HTTPException(status_code=400, detail="question cannot be empty")

    question_embedding = embed_texts([req.question])[0]
    results = vector_query(req.business_id, question_embedding, top_k=4)

    documents = results.get("documents", [[]])[0]
    metadatas = results.get("metadatas", [[]])[0]
    sources = sorted({m.get("title", "unknown") for m in metadatas}) if metadatas else []

    try:
        llm_result = generate_answer(req.question, documents)
    except Exception as e:
        raise HTTPException(status_code=502, detail=f"LLM generation failed: {str(e)}")

    return QueryResponse(
        answer=llm_result["answer"],
        sources=sources,
        session_id=req.session_id,
        tokens_used=llm_result["tokens_used"],
    )


@app.delete("/documents/{doc_id}")
def delete_document(doc_id: str, business_id: str):
    delete_by_doc_id(business_id, doc_id)
    return {"message": "deleted", "doc_id": doc_id}
