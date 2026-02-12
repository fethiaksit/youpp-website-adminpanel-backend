import Head from 'next/head';
import { useState } from 'react';
import { API_BASE } from '../lib/apiClient';
import styles from '../styles/Admin.module.css';

export default function Login() {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');

  async function onSubmit(e) {
    e.preventDefault();
    const res = await fetch(`${API_BASE}/api/auth/login`, {
      method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ email, password })
    });
    if (!res.ok) return alert('Login failed');
    const data = await res.json();
    localStorage.setItem('accessToken', data.accessToken);
    localStorage.setItem('refreshToken', data.refreshToken);
    window.location.href = '/admin';
  }

  return (
    <>
      <Head>
        <title>Super Admin Login | Youpp</title>
        <meta
          name='description'
          content='Login to the Youpp Super Admin panel to manage users, sites, and administrative operations.'
        />
        <meta name='robots' content='index,follow' />
      </Head>
      <div className={styles.container}>
        <h1>Login</h1>
        <p>Use your Super Admin credentials to access protected admin operations.</p>
        <form onSubmit={onSubmit}>
          <div className={styles.formRow}>
            <input className={styles.input} placeholder='Email' value={email} onChange={e=>setEmail(e.target.value)} />
          </div>
          <div className={styles.formRow}>
            <input className={styles.input} type='password' placeholder='Password' value={password} onChange={e=>setPassword(e.target.value)} />
          </div>
          <button className={styles.button}>Login</button>
        </form>
      </div>
    </>
  );
}
