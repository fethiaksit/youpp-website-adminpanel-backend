import Head from 'next/head';
import Link from 'next/link';
import { useEffect, useState } from 'react';
import { apiFetch, isAuthenticated } from '../../../lib/apiClient';
import styles from '../../../styles/Admin.module.css';

export default function SitesPage() {
  const [sites, setSites] = useState([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    async function loadSites() {
      if (!isAuthenticated()) {
        window.location.href = '/login';
        return;
      }
      const res = await apiFetch('/api/sites');
      if (!res.ok) {
        setLoading(false);
        return;
      }
      setSites(await res.json());
      setLoading(false);
    }
    loadSites();
  }, []);

  if (loading) return <div className={styles.container}>Yükleniyor...</div>;

  return (
    <>
      <Head><title>Siteler | Youpp Panel</title></Head>
      <div className={styles.container}>
        <h1>Sitelerim</h1>
        <table className={styles.table}>
          <thead>
            <tr>
              <th>Ad</th>
              <th>Slug</th>
              <th>Durum</th>
              <th>İşlem</th>
            </tr>
          </thead>
          <tbody>
            {sites.map((site) => (
              <tr key={site.id}>
                <td>{site.name}</td>
                <td>{site.slug}</td>
                <td>{site.status}</td>
                <td><Link className={styles.link} href={`/admin/sites/${site.id}`}>Düzenle</Link></td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </>
  );
}
