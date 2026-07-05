"""
Persistent ChromaDB vector store. Each business gets its own collection
(named after its business_id) so one business's documents are never
retrievable when answering another business's customers.
"""

import os
import chromadb

_client = None

CHROMA_PATH = os.getenv("CHROMA_PATH", "/data/chroma")


def get_client():
    global _client
    if _client is None:
        os.makedirs(CHROMA_PATH, exist_ok=True)
        _client = chromadb.PersistentClient(path=CHROMA_PATH)
    return _client


def collection_name(business_id: str) -> str:
    # Chroma collection names have character restrictions; keep it simple.
    return f"biz_{business_id}"


def get_collection(business_id: str):
    client = get_client()
    return client.get_or_create_collection(name=collection_name(business_id))


def add_chunks(business_id: str, doc_id: str, title: str, chunks: list[str], embeddings: list[list[float]]):
    collection = get_collection(business_id)
    ids = [f"{doc_id}_{i}" for i in range(len(chunks))]
    metadatas = [{"doc_id": doc_id, "title": title, "chunk_index": i} for i in range(len(chunks))]
    collection.add(ids=ids, embeddings=embeddings, documents=chunks, metadatas=metadatas)


def query(business_id: str, question_embedding: list[float], top_k: int = 4):
    collection = get_collection(business_id)
    if collection.count() == 0:
        return {"documents": [[]], "metadatas": [[]]}
    return collection.query(query_embeddings=[question_embedding], n_results=min(top_k, collection.count()))


def delete_by_doc_id(business_id: str, doc_id: str):
    collection = get_collection(business_id)
    collection.delete(where={"doc_id": doc_id})
