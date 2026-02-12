import { useRouter } from 'next/router';
import { useEffect, useState } from 'react';
import { apiFetch } from '../../../../lib/apiClient';
import styles from '../../../../styles/Admin.module.css';

export default function AccessPage() {
  const { query } = useRouter();
  const [email, setEmail] = useState('');
  const [role, setRole] = useState('viewer');
  const [users, setUsers] = useState([]);

  async function load() {
    if (!query.id) return;
    const res = await apiFetch(`/api/admin/sites/${query.id}/users`);
    setUsers(await res.json());
  }
  useEffect(() => { load(); }, [query.id]);

  async function grant(e) {
    e.preventDefault();
    const res = await apiFetch(`/api/admin/sites/${query.id}/grant`, { method: 'POST', body: JSON.stringify({ email, role }) });
    if (!res.ok) return alert('Grant failed');
    setEmail(''); load();
  }

  return <div className={styles.container}><h1>Site Access</h1><form onSubmit={grant} className={styles.card}><div className={styles.formRow}><input className={styles.input} placeholder='User email' value={email} onChange={e=>setEmail(e.target.value)} /><select className={styles.select} value={role} onChange={e=>setRole(e.target.value)}><option value='owner'>owner</option><option value='editor'>editor</option><option value='viewer'>viewer</option></select><button className={styles.button}>Grant</button></div></form><table className={styles.table}><thead><tr><th>Email</th><th>Role</th><th>Global Role</th></tr></thead><tbody>{users.map((u,i)=><tr key={i}><td>{u.email}</td><td>{u.role}</td><td>{u.globalRole}</td></tr>)}</tbody></table></div>;
}
