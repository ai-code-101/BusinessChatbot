"""
Calls GitHub Models' OpenAI-compatible chat completions endpoint to
generate the final answer, using retrieved chunks as context. Swapping
this out later for Claude/OpenAI directly only means changing this file -
nothing else in the pipeline needs to know.
"""

import os
import requests

GITHUB_MODELS_URL = "https://models.github.ai/inference/chat/completions"


def generate_answer(question: str, context_chunks: list[str], model: str = None) -> dict:
    token = os.getenv("GITHUB_TOKEN")
    if not token:
        raise RuntimeError("GITHUB_TOKEN environment variable is not set")

    model = model or os.getenv("GITHUB_MODEL", "openai/gpt-4.1-mini")

    context_text = "\n\n---\n\n".join(context_chunks) if context_chunks else "No relevant information found."

    system_prompt = (
        "You are a helpful assistant answering questions about a specific business, "
        "using only the context provided below. If the answer isn't in the context, "
        "say you don't have that information rather than guessing.\n\n"
        f"Context:\n{context_text}"
    )

    payload = {
        "model": model,
        "messages": [
            {"role": "system", "content": system_prompt},
            {"role": "user", "content": question},
        ],
        "temperature": 0.3,
    }

    headers = {
        "Authorization": f"Bearer {token}",
        "Content-Type": "application/json",
    }

    response = requests.post(GITHUB_MODELS_URL, json=payload, headers=headers, timeout=60)
    response.raise_for_status()
    data = response.json()

    answer = data["choices"][0]["message"]["content"]
    usage = data.get("usage", {})
    total_tokens = usage.get("total_tokens", 0)

    return {"answer": answer, "tokens_used": total_tokens}
