const API_BASE = process.env.NEXT_PUBLIC_API_BASE_URL || 'https://api.youpp.com.tr';

function getTokens() {
  if (typeof window === 'undefined') {
    return { accessToken: '', refreshToken: '' };
  }
  return {
    accessToken: localStorage.getItem('accessToken'),
    refreshToken: localStorage.getItem('refreshToken'),
  };
}

function clearTokens() {
  if (typeof window === 'undefined') return;
  localStorage.removeItem('accessToken');
  localStorage.removeItem('refreshToken');
}

async function refreshToken() {
  const { refreshToken: token } = getTokens();
  if (!token) throw new Error('No refresh token');

  const res = await fetch(`${API_BASE}/api/auth/refresh`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ refreshToken: token }),
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
  if (res.status !== 401 || path === '/api/auth/refresh') {
    return res;
  }

  try {
    const newToken = await refreshToken();
    headers.Authorization = `Bearer ${newToken}`;
    res = await fetch(`${API_BASE}${path}`, { ...options, headers });
    if (res.status !== 401) return res;
  } catch (_) {
    // ignored on purpose; handled below
  }

  clearTokens();
  if (typeof window !== 'undefined') {
    window.location.href = '/login';
  }
  return res;
}

export function isAuthenticated() {
  if (typeof window === 'undefined') return false;
  return Boolean(localStorage.getItem('accessToken'));
}

export { API_BASE, clearTokens };
