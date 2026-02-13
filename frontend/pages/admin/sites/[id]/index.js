import { useRouter } from 'next/router';
import { useEffect, useState } from 'react';
import { apiFetch, isAuthenticated } from '../../../../lib/apiClient';
import styles from '../../../../styles/Admin.module.css';

export default function SiteDetails() {
  const router = useRouter();
  const [site, setSite] = useState(null);
  const [content, setContent] = useState('{}');
  const [error, setError] = useState('');

  async function load() {
    if (!router.query.id) return;
    const res = await apiFetch(`/api/sites/${router.query.id}`);
    if (!res.ok) {
      setError('Site getirilemedi.');
      return;
    }
    const data = await res.json();
    setSite(data);
    setContent(JSON.stringify(data.content || {}, null, 2));
  }

  useEffect(() => {
    if (!isAuthenticated()) {
      window.location.href = '/login';
      return;
    }
    load();
  }, [router.query.id]);

  async function save() {
    try {
      const parsed = JSON.parse(content);
      const res = await apiFetch(`/api/sites/${router.query.id}/content`, {
        method: 'PUT',
        body: JSON.stringify({ content: parsed }),
      });
      if (!res.ok) {
        setError('Kaydetme başarısız.');
        return;
      }
      setError('');
      load();
    } catch (_) {
      setError('Geçerli JSON girin.');
    }
  }

  async function toggle(publish) {
    const endpoint = publish ? 'publish' : 'unpublish';
    const res = await apiFetch(`/api/sites/${router.query.id}/${endpoint}`, { method: 'POST' });
    if (!res.ok) {
      setError('Durum güncellenemedi.');
      return;
    }
    setError('');
    load();
  }

  if (!site) return <div className={styles.container}>Yükleniyor...</div>;

  return (
    <div className={styles.container}>
      <h1>{site.name}</h1>
      <p>Slug: {site.slug} | Durum: {site.status}</p>
      <textarea className={styles.textarea} rows={20} value={content} onChange={(e) => setContent(e.target.value)} />
      {error ? <p className={styles.error}>{error}</p> : null}
      <div className={styles.formRow}>
        <button className={styles.button} onClick={save}>Kaydet</button>
        <button className={styles.button} onClick={() => toggle(true)}>Yayınla</button>
        <button className={styles.button} onClick={() => toggle(false)}>Yayından Kaldır</button>
      </div>
    </div>
  );
}
