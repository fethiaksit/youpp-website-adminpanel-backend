import { useRouter } from 'next/router';
import { useEffect, useState } from 'react';
import { apiFetch } from '../../../../lib/apiClient';
import styles from '../../../../styles/Admin.module.css';

export default function SiteDetails() {
  const { query } = useRouter();
  const [site, setSite] = useState(null);
  const [content, setContent] = useState('{}');

  async function load() {
    if (!query.id) return;
    const res = await apiFetch(`/api/sites/${query.id}`);
    const data = await res.json();
    setSite(data);
    setContent(JSON.stringify(data.content || {}, null, 2));
  }
  useEffect(() => { load(); }, [query.id]);

  async function save() {
    await apiFetch(`/api/sites/${query.id}/content`, { method: 'PUT', body: JSON.stringify({ content: JSON.parse(content) }) });
    load();
  }

  async function toggle(publish) {
    await apiFetch(`/api/sites/${query.id}/${publish ? 'publish' : 'unpublish'}`, { method: 'POST' });
    load();
  }

  if (!site) return <div className={styles.container}>Loading...</div>;
  return <div className={styles.container}><h1>{site.name}</h1><textarea className={styles.textarea} rows={20} value={content} onChange={e=>setContent(e.target.value)} /><div className={styles.formRow}><button className={styles.button} onClick={save}>Save</button><button className={styles.button} onClick={()=>toggle(true)}>Publish</button><button className={styles.button} onClick={()=>toggle(false)}>Unpublish</button></div></div>;
}
