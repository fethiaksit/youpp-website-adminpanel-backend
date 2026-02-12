import Head from 'next/head';
import { useRouter } from 'next/router';
import { useEffect, useState } from 'react';
import Link from 'next/link';
import { apiFetch, isAuthenticated } from '../../../lib/apiClient';
import styles from '../../../styles/Admin.module.css';

export default function SitesPage() {
  const router = useRouter();
  const [sites, setSites] = useState([]);
  const [name, setName] = useState('');
  const [slug, setSlug] = useState('');
  const [ready, setReady] = useState(false);

  async function load() {
    const res = await apiFetch('/api/sites');
    setSites(await res.json());
  }
  useEffect(() => {
    if (!isAuthenticated()) {
      router.replace('/login');
      return;
    }
    setReady(true);
    load();
  }, []);

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

  if (!ready) return <div className={styles.container}>Redirecting to login...</div>;

  return (
    <>
      <Head>
        <title>Admin Sites | Youpp</title>
        <meta name='robots' content='index,follow' />
      </Head>
      <div className={styles.container}><h1>Sites</h1><form onSubmit={createSite} className={styles.card}><div className={styles.formRow}><input className={styles.input} placeholder='Name' value={name} onChange={e=>setName(e.target.value)} /><input className={styles.input} placeholder='Slug' value={slug} onChange={e=>setSlug(e.target.value)} /><button className={styles.button}>Create</button></div></form><table className={styles.table}><thead><tr><th>Name</th><th>Slug</th><th>Status</th><th>Actions</th></tr></thead><tbody>{sites.map(s=><tr key={s.id}><td>{s.name}</td><td>{s.slug}</td><td>{s.status}</td><td><Link className={styles.link} href={`/admin/sites/${s.id}`}>Edit</Link> | <Link className={styles.link} href={`/admin/sites/${s.id}/access`}>Access</Link> | <button className={styles.button} onClick={()=>toggle(s, s.status!=='published')}>{s.status==='published'?'Unpublish':'Publish'}</button></td></tr>)}</tbody></table></div>
    </>
  );
}
