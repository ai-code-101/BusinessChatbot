import { useEffect, useRef, useState } from "react";
import "./ChatWidget.css";

const API_URL = import.meta.env.VITE_API_URL || "http://localhost:8080";

// One session id per browser tab load, so the backend could (later) thread
// multi-turn context together. Regenerating on refresh is fine for now.
const sessionId =
  typeof crypto !== "undefined" && crypto.randomUUID
    ? crypto.randomUUID()
    : `sess_${Date.now()}`;

export default function ChatWidget({ businessId, businessName }) {
  const [open, setOpen] = useState(false);
  const [messages, setMessages] = useState([
    {
      role: "assistant",
      content: `Hi! I'm the ${businessName} assistant. Ask me anything about our services.`,
    },
  ]);
  const [input, setInput] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const scrollRef = useRef(null);

  useEffect(() => {
    if (scrollRef.current) {
      scrollRef.current.scrollTop = scrollRef.current.scrollHeight;
    }
  }, [messages, loading]);

  async function sendMessage(e) {
    e.preventDefault();
    const question = input.trim();
    if (!question || loading) return;

    setMessages((prev) => [...prev, { role: "user", content: question }]);
    setInput("");
    setLoading(true);
    setError(null);

    try {
      const res = await fetch(`${API_URL}/v1/chat/ask`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          business_id: businessId,
          question,
          session_id: sessionId,
        }),
      });

      if (!res.ok) {
        const body = await res.json().catch(() => ({}));
        throw new Error(body.error || "Something went wrong. Please try again.");
      }

      const data = await res.json();
      setMessages((prev) => [
        ...prev,
        { role: "assistant", content: data.answer, sources: data.sources },
      ]);
    } catch (err) {
      setError(err.message || "Could not reach the assistant.");
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="cw-root">
      {open && (
        <div className="cw-panel" role="dialog" aria-label={`Chat with ${businessName}`}>
          <div className="cw-header">
            <div className="cw-header-accent" aria-hidden="true" />
            <div className="cw-header-content">
              <div className="cw-avatar">{businessName.charAt(0)}</div>
              <div className="cw-header-text">
                <div className="cw-header-name">{businessName}</div>
                <div className="cw-header-status">
                  <span className="cw-status-dot" />
                  Online
                </div>
              </div>
              <button
                className="cw-close"
                onClick={() => setOpen(false)}
                aria-label="Close chat"
              >
                ×
              </button>
            </div>
          </div>

          <div className="cw-messages" ref={scrollRef}>
            {messages.map((m, i) => (
              <div key={i} className={`cw-message cw-message-${m.role}`}>
                <div className="cw-bubble">{m.content}</div>
                {m.sources && m.sources.length > 0 && (
                  <div className="cw-sources">Source: {m.sources.join(", ")}</div>
                )}
              </div>
            ))}
            {loading && (
              <div className="cw-message cw-message-assistant">
                <div className="cw-bubble cw-typing">
                  <span />
                  <span />
                  <span />
                </div>
              </div>
            )}
            {error && <div className="cw-error">{error}</div>}
          </div>

          <form className="cw-input-row" onSubmit={sendMessage}>
            <input
              type="text"
              value={input}
              onChange={(e) => setInput(e.target.value)}
              placeholder="Type your question..."
              disabled={loading}
              className="cw-input"
            />
            <button type="submit" className="cw-send" disabled={loading || !input.trim()}>
              <svg viewBox="0 0 24 24" width="18" height="18" fill="none">
                <path
                  d="M3 11.5L21 3l-7.5 18-3-7.5L3 11.5z"
                  stroke="currentColor"
                  strokeWidth="1.8"
                  strokeLinejoin="round"
                  strokeLinecap="round"
                />
              </svg>
            </button>
          </form>
        </div>
      )}

      <button
        className={`cw-launcher ${open ? "cw-launcher-open" : ""}`}
        onClick={() => setOpen((v) => !v)}
        aria-label={open ? "Close chat" : "Chat with us"}
      >
        {open ? (
          <span className="cw-launcher-x">×</span>
        ) : (
          <span className="cw-launcher-icon" aria-hidden="true" />
        )}
      </button>
    </div>
  );
}
