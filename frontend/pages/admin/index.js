import Head from 'next/head';
import Link from 'next/link';
import { useEffect, useState } from 'react';
import { apiFetch, clearTokens, isAuthenticated } from '../../lib/apiClient';
import styles from '../../styles/Admin.module.css';

export default function AdminHome() {
  const [me, setMe] = useState(null);

  useEffect(() => {
    async function loadMe() {
      if (!isAuthenticated()) {
        window.location.href = '/login';
        return;
      }
      const res = await apiFetch('/api/me');
      if (!res.ok) return;
      setMe(await res.json());
    }
    loadMe();
  }, []);

  if (!me) {
    return <div className={styles.container}>Yükleniyor...</div>;
  }

  const isSuperAdmin = me.globalRole === 'superadmin';

  return (
    <>
      <Head><title>Admin | Youpp Panel</title></Head>
      <div className={styles.container}>
        <div className={styles.header}>
          <div>
            <h1>Admin Panel</h1>
            <p>{me.email}</p>
          </div>
          <button className={styles.button} onClick={() => { clearTokens(); window.location.href = '/login'; }}>Çıkış</button>
        </div>

        <div className={styles.card}>
          <ul>
            <li><Link className={styles.link} href='/admin/sites'>Sitelerim</Link></li>
            {isSuperAdmin ? <li><Link className={styles.link} href='/admin/users'>Kullanıcılar</Link></li> : null}
            {isSuperAdmin ? <li><Link className={styles.link} href='/admin/sites'>Site Erişim Yönetimi</Link></li> : null}
          </ul>
        </div>
      </div>
    </>
  );
}
