"""
Wraps a local sentence-transformers model so we can turn text chunks and
questions into vectors without any external API calls (fast, free, and
runs fine on CPU for this model size).
"""

from sentence_transformers import SentenceTransformer

_model = None


def get_model():
    global _model
    if _model is None:
        # Small, fast, good enough for FAQ/knowledge-base style retrieval.
        _model = SentenceTransformer("all-MiniLM-L6-v2")
    return _model


def embed_texts(texts: list[str]) -> list[list[float]]:
    model = get_model()
    embeddings = model.encode(texts, convert_to_numpy=True)
    return embeddings.tolist()
