import { useEffect, useState } from 'react';
import Link from 'next/link';
import { apiFetch } from '../../../lib/apiClient';
import styles from '../../../styles/Admin.module.css';

export default function SitesPage() {
  const [sites, setSites] = useState([]);
  const [name, setName] = useState('');
  const [slug, setSlug] = useState('');

  async function load() {
    const res = await apiFetch('/api/sites');
    setSites(await res.json());
  }
  useEffect(() => { load(); }, []);

  async function createSite(e) {
    e.preventDefault();
    const res = await apiFetch('/api/admin/sites', { method: 'POST', body: JSON.stringify({ name, slug }) });
    if (!res.ok) return alert('Create failed');
    setName(''); setSlug(''); load();
  }

  async function toggle(site, publish) {
    await apiFetch(`/api/sites/${site.id}/` + (publish ? 'publish' : 'unpublish'), { method: 'POST' });
    load();
  }

  return <div className={styles.container}><h1>Sites</h1><form onSubmit={createSite} className={styles.card}><div className={styles.formRow}><input className={styles.input} placeholder='Name' value={name} onChange={e=>setName(e.target.value)} /><input className={styles.input} placeholder='Slug' value={slug} onChange={e=>setSlug(e.target.value)} /><button className={styles.button}>Create</button></div></form><table className={styles.table}><thead><tr><th>Name</th><th>Slug</th><th>Status</th><th>Actions</th></tr></thead><tbody>{sites.map(s=><tr key={s.id}><td>{s.name}</td><td>{s.slug}</td><td>{s.status}</td><td><Link className={styles.link} href={`/admin/sites/${s.id}`}>Edit</Link> | <Link className={styles.link} href={`/admin/sites/${s.id}/access`}>Access</Link> | <button className={styles.button} onClick={()=>toggle(s, s.status!=='published')}>{s.status==='published'?'Unpublish':'Publish'}</button></td></tr>)}</tbody></table></div>;
}
