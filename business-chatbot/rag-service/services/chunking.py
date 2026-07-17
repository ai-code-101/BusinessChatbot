"""
Splits raw text (from an uploaded file or pasted text) into overlapping
chunks small enough to embed and retrieve individually. Overlap keeps
context from being cut off awkwardly at chunk boundaries.
"""

def chunk_text(text: str, chunk_size: int = 300, overlap: int = 35) -> list[str]:
    text = text.strip()
    if not text:
        return []

    chunks = []
    start = 0
    text_len = len(text)

    while start < text_len:
        end = min(start + chunk_size, text_len)
        chunk = text[start:end].strip()
        if chunk:
            chunks.append(chunk)
        if end == text_len:
            break
        start = end - overlap  # step back a bit so context overlaps

    return chunks
