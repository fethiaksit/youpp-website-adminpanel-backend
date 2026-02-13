import Head from 'next/head';
import { useState } from 'react';
import { API_BASE } from '../lib/apiClient';
import styles from '../styles/Register.module.css';

export default function RegisterPage() {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  async function onSubmit(e) {
    e.preventDefault();
    setError('');
    setLoading(true);

    try {
      const res = await fetch(`${API_BASE}/api/public/register`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email, password }),
      });

      const data = await res.json().catch(() => ({}));
      if (!res.ok) {
        setError(data.error || 'Kayıt sırasında bir hata oluştu.');
        return;
      }

      localStorage.setItem('accessToken', data.accessToken);
      localStorage.setItem('refreshToken', data.refreshToken);
      window.location.href = 'https://panel.youpp.com.tr/admin';
    } finally {
      setLoading(false);
    }
  }

  return (
    <>
      <Head>
        <title>Kayıt Ol | Youpp</title>
      </Head>
      <main className={styles.page}>
        <form className={styles.form} onSubmit={onSubmit}>
          <h1>Ücretsiz Başla</h1>
          <label className={styles.field}>
            <span>E-posta</span>
            <input type='email' value={email} onChange={(e) => setEmail(e.target.value)} required />
          </label>
          <label className={styles.field}>
            <span>Şifre (min. 8 karakter)</span>
            <input type='password' value={password} onChange={(e) => setPassword(e.target.value)} minLength={8} required />
          </label>
          {error ? <p className={styles.error}>{error}</p> : null}
          <button className={styles.button} disabled={loading}>{loading ? 'Kaydediliyor...' : 'Hesap Oluştur'}</button>
        </form>
      </main>
    </>
  );
}
