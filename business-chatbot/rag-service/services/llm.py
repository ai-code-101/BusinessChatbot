"""
Multi-provider LLM router. Supports both Anthropic and GitHub Models behind
one function, selected by a simple model key stored in Mongo (via the admin
settings endpoint) rather than hardcoded in an env var. Adding a new model
later just means adding one entry to AVAILABLE_MODELS below.
"""

import os
import requests

ANTHROPIC_URL = "https://api.anthropic.com/v1/messages"
GITHUB_MODELS_URL = "https://models.github.ai/inference/chat/completions"

# Keys here are what the admin picks from in the dashboard dropdown.
# "id" is the actual model identifier sent to that provider's API.
AVAILABLE_MODELS = {
    "claude-haiku-4-5": {"provider": "anthropic", "id": "claude-haiku-4-5-20251001"},
    "claude-sonnet-5": {"provider": "anthropic", "id": "claude-sonnet-5"},
    "gpt-4.1-mini": {"provider": "github", "id": "openai/gpt-4.1-mini"},
    "llama-4": {"provider": "github", "id": "meta/llama-4"},
}

DEFAULT_MODEL_KEY = "claude-haiku-4-5"

def _build_system_prompt(context_chunks: list[str]) -> str:
    context_text = "\n---\n".join(context_chunks) if context_chunks else "No relevant information found."
    return (
        "Answer using only this context. If it's not covered, say you don't know. "
        "Do not make up answers."
        "Simplify and summarize the context as much as possible, but do not leave out any relevant information. "
        "after giving your response if necesssary you can the give a follow up question to the user to clarify their request. "
         "Be concise, but format for readability: use numbered lists for steps or multiple "
        "items, bullet points for unordered items, and **bold** for key terms. Put each list "
        "item on its own line - never combine multiple list items into one paragraph. "
        "Only use structure when the content actually has multiple items; a short factual "
        "Answer with a maxximum of 2 sentenses for answers that need satting just satte them without giving any description" 
    "answers should be formated as a list with numbers in its own line when needed, and dont show the user that you are not a human speak like a profassional business person."
        f"Context:\n{context_text}"
    )


def _call_anthropic(model_id: str, question: str, system_prompt: str, max_tokens: int) -> dict:
    api_key = os.getenv("ANTHROPIC_API_KEY")
    if not api_key:
        raise RuntimeError("ANTHROPIC_API_KEY environment variable is not set")

    payload = {
        "model": model_id,
        "max_tokens": max_tokens,
        "system": system_prompt,
        "messages": [{"role": "user", "content": question}],
        "temperature": 0.3,
    }
    headers = {
        "x-api-key": api_key,
        "anthropic-version": "2023-06-01",
        "Content-Type": "application/json",
    }

    response = requests.post(ANTHROPIC_URL, json=payload, headers=headers, timeout=60)
    response.raise_for_status()
    data = response.json()

    answer = data["content"][0]["text"]
    usage = data.get("usage", {})
    total_tokens = usage.get("input_tokens", 0) + usage.get("output_tokens", 0)
    return {"answer": answer, "tokens_used": total_tokens}


def _call_github_models(model_id: str, question: str, system_prompt: str, max_tokens: int) -> dict:
    token = os.getenv("GITHUB_TOKEN")
    if not token:
        raise RuntimeError("GITHUB_TOKEN environment variable is not set")

    payload = {
        "model": model_id,
        "messages": [
            {"role": "system", "content": system_prompt},
            {"role": "user", "content": question},
        ],
        "temperature": 0.3,
        "max_tokens": max_tokens,
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


def generate_answer(question: str, context_chunks: list[str], model_key: str = None) -> dict:
    model_key = model_key or os.getenv("DEFAULT_MODEL_KEY", DEFAULT_MODEL_KEY)
    model_info = AVAILABLE_MODELS.get(model_key)

    if not model_info:
        raise RuntimeError(f"Unknown model key: {model_key}")

    max_output_tokens = int(os.getenv("MAX_ANSWER_TOKENS", "250"))
    system_prompt = _build_system_prompt(context_chunks)

    if model_info["provider"] == "anthropic":
        result = _call_anthropic(model_info["id"], question, system_prompt, max_output_tokens)
    elif model_info["provider"] == "github":
        result = _call_github_models(model_info["id"], question, system_prompt, max_output_tokens)
    else:
        raise RuntimeError(f"Unknown provider for model: {model_key}")

    result["model_key"] = model_key
    return result
