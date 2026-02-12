const API_BASE = process.env.NEXT_PUBLIC_API_BASE || 'http://localhost:8080';

function getTokens() {
  return {
    accessToken: localStorage.getItem('accessToken'),
    refreshToken: localStorage.getItem('refreshToken'),
  };
}

async function refreshToken() {
  const { refreshToken } = getTokens();
  if (!refreshToken) throw new Error('No refresh token');
  const res = await fetch(`${API_BASE}/api/auth/refresh`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ refreshToken }),
  });
  if (!res.ok) throw new Error('Unable to refresh token');
  const data = await res.json();
  localStorage.setItem('accessToken', data.accessToken);
  localStorage.setItem('refreshToken', data.refreshToken);
  return data.accessToken;
}

export async function apiFetch(path, options = {}) {
  const { accessToken } = getTokens();
  const headers = {
    'Content-Type': 'application/json',
    ...(options.headers || {}),
  };
  if (accessToken) headers.Authorization = `Bearer ${accessToken}`;

  let res = await fetch(`${API_BASE}${path}`, { ...options, headers });
  if (res.status === 401 && path !== '/api/auth/refresh') {
    const newToken = await refreshToken();
    headers.Authorization = `Bearer ${newToken}`;
    res = await fetch(`${API_BASE}${path}`, { ...options, headers });
  }
  return res;
}

export function isAuthenticated() {
  if (typeof window === 'undefined') return false;
  return Boolean(localStorage.getItem('accessToken'));
}

export { API_BASE };
