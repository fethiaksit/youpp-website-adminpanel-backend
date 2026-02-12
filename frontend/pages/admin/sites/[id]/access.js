import Head from 'next/head';
import { useRouter } from 'next/router';
import { useEffect, useState } from 'react';
import { apiFetch, isAuthenticated } from '../../../../lib/apiClient';
import styles from '../../../../styles/Admin.module.css';

export default function AccessPage() {
  const router = useRouter();
  const { query } = router;
  const [email, setEmail] = useState('');
  const [role, setRole] = useState('viewer');
  const [users, setUsers] = useState([]);
  const [ready, setReady] = useState(false);

  async function load() {
    if (!query.id) return;
    const res = await apiFetch(`/api/admin/sites/${query.id}/users`);
    setUsers(await res.json());
  }
  useEffect(() => {
    if (!isAuthenticated()) {
      router.replace('/login');
      return;
    }
    setReady(true);
    load();
  }, [query.id]);

  async function grant(e) {
    e.preventDefault();
    const res = await apiFetch(`/api/admin/sites/${query.id}/grant`, { method: 'POST', body: JSON.stringify({ email, role }) });
    if (!res.ok) return alert('Grant failed');
    setEmail(''); load();
  }

  if (!ready) return <div className={styles.container}>Redirecting to login...</div>;

  return (
    <>
      <Head>
        <title>Admin Site Access | Youpp</title>
        <meta name='robots' content='index,follow' />
      </Head>
      <div className={styles.container}><h1>Site Access</h1><form onSubmit={grant} className={styles.card}><div className={styles.formRow}><input className={styles.input} placeholder='User email' value={email} onChange={e=>setEmail(e.target.value)} /><select className={styles.select} value={role} onChange={e=>setRole(e.target.value)}><option value='owner'>owner</option><option value='editor'>editor</option><option value='viewer'>viewer</option></select><button className={styles.button}>Grant</button></div></form><table className={styles.table}><thead><tr><th>Email</th><th>Role</th><th>Global Role</th></tr></thead><tbody>{users.map((u,i)=><tr key={i}><td>{u.email}</td><td>{u.role}</td><td>{u.globalRole}</td></tr>)}</tbody></table></div>
    </>
  );
}
