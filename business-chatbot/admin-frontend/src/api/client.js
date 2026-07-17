const API_URL = import.meta.env.VITE_API_URL || "http://localhost:8080";

function getToken() {
  return localStorage.getItem("admin_token");
}

function authHeaders() {
  const token = getToken();
  return token ? { Authorization: `Bearer ${token}` } : {};
}

async function handleResponse(res) {
  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    throw new Error(body.error || `Request failed (${res.status})`);
  }
  return res.json();
}

export async function login(email, password) {
  const res = await fetch(`${API_URL}/v1/admin/login`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ email, password }),
  });
  const data = await handleResponse(res);
  localStorage.setItem("admin_token", data.token);
  localStorage.setItem("admin_email", data.email);
  localStorage.setItem("admin_business_id", data.business_id);
  return data;
}

export function logout() {
  localStorage.removeItem("admin_token");
  localStorage.removeItem("admin_email");
  localStorage.removeItem("admin_business_id");
}

export function isAuthenticated() {
  return Boolean(getToken());
}

export function currentAdmin() {
  return {
    email: localStorage.getItem("admin_email"),
    businessId: localStorage.getItem("admin_business_id"),
  };
}

export async function fetchDocuments() {
  const res = await fetch(`${API_URL}/v1/admin/documents`, {
    headers: { ...authHeaders() },
  });
  const data = await handleResponse(res);
  return data.documents || [];
}

export async function uploadText(title, content) {
  const res = await fetch(`${API_URL}/v1/admin/documents/text`, {
    method: "POST",
    headers: { "Content-Type": "application/json", ...authHeaders() },
    body: JSON.stringify({ title, content }),
  });
  return handleResponse(res);
}

export async function uploadFile(file) {
  const formData = new FormData();
  formData.append("file", file);
  const res = await fetch(`${API_URL}/v1/admin/documents/upload`, {
    method: "POST",
    headers: { ...authHeaders() }, // no Content-Type - browser sets multipart boundary
    body: formData,
  });
  return handleResponse(res);
}

export async function deleteDocument(id) {
  const res = await fetch(`${API_URL}/v1/admin/documents/${id}`, {
    method: "DELETE",
    headers: { ...authHeaders() },
  });
  return handleResponse(res);
}

export async function fetchUsageSummary() {
  const res = await fetch(`${API_URL}/v1/admin/usage/summary`, {
    headers: { ...authHeaders() },
  });
  return handleResponse(res);
}

export async function fetchUsageLogs() {
  const res = await fetch(`${API_URL}/v1/admin/usage/logs`, {
    headers: { ...authHeaders() },
  });
  const data = await handleResponse(res);
  return data.logs || [];
}

export async function fetchModelSetting() {
  const res = await fetch(`${API_URL}/v1/admin/settings/model`, {
    headers: { ...authHeaders() },
  });
  return handleResponse(res);
}

export async function setModelSetting(modelKey) {
  const res = await fetch(`${API_URL}/v1/admin/settings/model`, {
    method: "PUT",
    headers: { "Content-Type": "application/json", ...authHeaders() },
    body: JSON.stringify({ model_key: modelKey }),
  });
  return handleResponse(res);
}
