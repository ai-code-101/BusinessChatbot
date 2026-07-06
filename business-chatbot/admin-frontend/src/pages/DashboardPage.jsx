import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import {
  currentAdmin,
  logout,
  fetchDocuments,
  uploadText,
  uploadFile,
  deleteDocument,
  fetchUsageSummary,
  fetchUsageLogs,
} from "../api/client.js";
import "./DashboardPage.css";

export default function DashboardPage() {
  const navigate = useNavigate();
  const admin = currentAdmin();

  const [documents, setDocuments] = useState([]);
  const [loadingDocs, setLoadingDocs] = useState(true);
  const [error, setError] = useState(null);

  const [usageSummary, setUsageSummary] = useState(null);
  const [usageLogs, setUsageLogs] = useState([]);
  const [loadingUsage, setLoadingUsage] = useState(true);

  const [mode, setMode] = useState("text"); // "text" | "file"
  const [title, setTitle] = useState("");
  const [content, setContent] = useState("");
  const [file, setFile] = useState(null);
  const [submitting, setSubmitting] = useState(false);
  const [formError, setFormError] = useState(null);

  async function loadDocuments() {
    setLoadingDocs(true);
    setError(null);
    try {
      const docs = await fetchDocuments();
      setDocuments(docs);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoadingDocs(false);
    }
  }

  async function loadUsage() {
    setLoadingUsage(true);
    try {
      const [summary, logs] = await Promise.all([fetchUsageSummary(), fetchUsageLogs()]);
      setUsageSummary(summary);
      setUsageLogs(logs);
    } catch (err) {
      console.error("Failed to load usage:", err.message);
    } finally {
      setLoadingUsage(false);
    }
  }

  useEffect(() => {
    loadDocuments();
    loadUsage();
  }, []);

  function handleLogout() {
    logout();
    navigate("/login");
  }

  async function handleSubmit(e) {
    e.preventDefault();
    setFormError(null);
    setSubmitting(true);
    try {
      if (mode === "text") {
        if (!title.trim() || !content.trim()) {
          throw new Error("Title and content are both required.");
        }
        await uploadText(title.trim(), content.trim());
      } else {
        if (!file) {
          throw new Error("Choose a file first.");
        }
        await uploadFile(file);
      }
      setTitle("");
      setContent("");
      setFile(null);
      await loadDocuments();
    } catch (err) {
      setFormError(err.message);
    } finally {
      setSubmitting(false);
    }
  }

  async function handleDelete(id) {
    if (!window.confirm("Delete this document? This can't be undone.")) return;
    try {
      await deleteDocument(id);
      await loadDocuments();
    } catch (err) {
      setError(err.message);
    }
  }

  return (
    <div className="dash-page">
      <header className="dash-header">
        <div className="dash-header-accent" aria-hidden="true" />
        <div className="dash-header-content">
          <div>
            <div className="dash-title">Knowledge Base</div>
            <div className="dash-subtitle">
              {admin.email} · {admin.businessId}
            </div>
          </div>
          <button className="dash-logout" onClick={handleLogout}>
            Log out
          </button>
        </div>
      </header>

      <main className="dash-main">
        <section className="dash-card">
          <h2 className="dash-card-title">Add knowledge</h2>
          <div className="dash-tabs">
            <button
              className={`dash-tab ${mode === "text" ? "dash-tab-active" : ""}`}
              onClick={() => setMode("text")}
              type="button"
            >
              Paste text
            </button>
            <button
              className={`dash-tab ${mode === "file" ? "dash-tab-active" : ""}`}
              onClick={() => setMode("file")}
              type="button"
            >
              Upload file
            </button>
          </div>

          <form onSubmit={handleSubmit} className="dash-form">
            {mode === "text" ? (
              <>
                <label className="dash-label">Title</label>
                <input
                  className="dash-input"
                  value={title}
                  onChange={(e) => setTitle(e.target.value)}
                  placeholder="e.g. Delivery Policy"
                />
                <label className="dash-label">Content</label>
                <textarea
                  className="dash-textarea"
                  value={content}
                  onChange={(e) => setContent(e.target.value)}
                  placeholder="Paste the information customers should be able to ask about..."
                  rows={6}
                />
              </>
            ) : (
              <>
                <label className="dash-label">Text file</label>
                <input
                  type="file"
                  accept=".txt"
                  className="dash-file-input"
                  onChange={(e) => setFile(e.target.files[0] || null)}
                />
                <p className="dash-hint">Plain text (.txt) files only, for now.</p>
              </>
            )}

            {formError && <div className="dash-error">{formError}</div>}

            <button type="submit" className="dash-submit" disabled={submitting}>
              {submitting ? "Processing..." : "Add to knowledge base"}
            </button>
          </form>
        </section>

        <section className="dash-card">
          <h2 className="dash-card-title">Existing documents</h2>

          {loadingDocs && <p className="dash-hint">Loading...</p>}
          {error && <div className="dash-error">{error}</div>}

          {!loadingDocs && documents.length === 0 && (
            <p className="dash-hint">
              Nothing here yet. Add your first document above so the chatbot has
              something to answer questions from.
            </p>
          )}

          <ul className="dash-doc-list">
            {documents.map((doc) => (
              <li key={doc.id} className="dash-doc-item">
                <div className="dash-doc-info">
                  <div className="dash-doc-title">{doc.title}</div>
                  <div className="dash-doc-meta">
                    {doc.source_type} · {doc.chunk_count} chunk
                    {doc.chunk_count !== 1 ? "s" : ""}
                  </div>
                  <div className="dash-doc-preview">{doc.preview}</div>
                </div>
                <button
                  className="dash-doc-delete"
                  onClick={() => handleDelete(doc.id)}
                  aria-label={`Delete ${doc.title}`}
                >
                  Delete
                </button>
              </li>
            ))}
          </ul>
        </section>

        <section className="dash-card dash-card-wide">
          <div className="dash-usage-header">
            <h2 className="dash-card-title">Token Usage</h2>
            <button className="dash-refresh" onClick={loadUsage} type="button">
              Refresh
            </button>
          </div>

          {loadingUsage && <p className="dash-hint">Loading usage...</p>}

          {!loadingUsage && usageSummary && (
            <div className="dash-usage-stats">
              <div className="dash-stat">
                <div className="dash-stat-value">{usageSummary.total_tokens.toLocaleString()}</div>
                <div className="dash-stat-label">Total tokens</div>
              </div>
              <div className="dash-stat">
                <div className="dash-stat-value">{usageSummary.total_messages.toLocaleString()}</div>
                <div className="dash-stat-label">Total questions asked</div>
              </div>
              <div className="dash-stat">
                <div className="dash-stat-value">{usageSummary.today_tokens.toLocaleString()}</div>
                <div className="dash-stat-label">Tokens today</div>
              </div>
              <div className="dash-stat">
                <div className="dash-stat-value">{usageSummary.today_messages.toLocaleString()}</div>
                <div className="dash-stat-label">Questions today</div>
              </div>
            </div>
          )}

          {!loadingUsage && usageLogs.length === 0 && (
            <p className="dash-hint">
              No conversations yet. Once customers start using the chat widget,
              every question and its token cost will show up here.
            </p>
          )}

          {!loadingUsage && usageLogs.length > 0 && (
            <div className="dash-usage-table-wrap">
              <table className="dash-usage-table">
                <thead>
                  <tr>
                    <th>When</th>
                    <th>Question</th>
                    <th>Tokens</th>
                  </tr>
                </thead>
                <tbody>
                  {usageLogs.map((log) => (
                    <tr key={log.id}>
                      <td className="dash-usage-time">
                        {new Date(log.created_at * 1000).toLocaleString()}
                      </td>
                      <td className="dash-usage-question">{log.question}</td>
                      <td className="dash-usage-tokens">{log.tokens_used}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </section>
      </main>
    </div>
  );
}