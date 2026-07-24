"""
Internal RAG microservice. This is never called directly by any frontend -
only by the Go backend. It handles:
  - /ingest   -> chunk + embed + store a document's text
  - /query    -> embed a question, retrieve relevant chunks, ask the LLM
  - /documents/{doc_id} (DELETE) -> remove a document's vectors
  - /onboarding/submit -> validate and email a captured name/phone
"""

from fastapi import FastAPI, HTTPException
from pydantic import BaseModel

from services.chunking import chunk_text
from services.embeddings import embed_texts
from services.vector_store import add_chunks, query as vector_query, delete_by_doc_id
from services.llm import generate_answer, AVAILABLE_MODELS
from services.onboarding import validate_details
from services.email_service import send_onboarding_email

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
    model_key: str = None


class QueryResponse(BaseModel):
    answer: str
    sources: list[str]
    session_id: str
    tokens_used: int
    model_key: str


class OnboardingRequest(BaseModel):
    business_id: str = ""
    session_id: str = ""
    name: str
    phone: str


class OnboardingResponse(BaseModel):
    success: bool
    message: str


@app.get("/health")
def health():
    return {"status": "ok"}


@app.get("/models")
def list_models():
    return {"models": list(AVAILABLE_MODELS.keys())}


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
    distances = results.get("distances", [[]])[0]

    documents, metadatas = filter_relevant(documents, metadatas, distances)

    sources = sorted({m.get("title", "unknown") for m in metadatas}) if metadatas else []

    try:
        llm_result = generate_answer(req.question, documents, req.model_key)
    except Exception as e:
        raise HTTPException(status_code=502, detail=f"LLM generation failed: {str(e)}")

    return QueryResponse(
        answer=llm_result["answer"],
        sources=sources,
        session_id=req.session_id,
        tokens_used=llm_result["tokens_used"],
        model_key=llm_result["model_key"],
    )


@app.post("/onboarding/submit", response_model=OnboardingResponse)
def onboarding_submit(req: OnboardingRequest):
    ok, error = validate_details(req.name, req.phone)
    if not ok:
        raise HTTPException(status_code=400, detail=error)

    try:
        send_onboarding_email(req.name, req.phone, req.business_id)
    except Exception as e:
        raise HTTPException(status_code=502, detail=f"Failed to send onboarding email: {str(e)}")

    return OnboardingResponse(success=True, message="Onboarding details received.")


def filter_relevant(documents, metadatas, distances, max_chunks=3, relative_tolerance=0.35):
    if not documents or not distances:
        return documents, metadatas

    best = distances[0]
    kept_docs, kept_meta = [], []
    for doc, meta, dist in zip(documents, metadatas, distances):
        if len(kept_docs) >= max_chunks:
            break
        if dist <= best * (1 + relative_tolerance) or len(kept_docs) == 0:
            kept_docs.append(doc)
            kept_meta.append(meta)
    return kept_docs, kept_meta


@app.delete("/documents/{doc_id}")
def delete_document(doc_id: str, business_id: str):
    delete_by_doc_id(business_id, doc_id)
    return {"message": "deleted", "doc_id": doc_id}