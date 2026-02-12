import { useEffect, useState } from 'react';
import { apiFetch } from '../../lib/apiClient';
import styles from '../../styles/Admin.module.css';

export default function UsersPage() {
  const [users, setUsers] = useState([]);
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [globalRole, setGlobalRole] = useState('user');

  async function load() {
    const res = await apiFetch('/api/admin/users');
    setUsers(await res.json());
  }
  useEffect(() => { load(); }, []);

  async function create(e) {
    e.preventDefault();
    const res = await apiFetch('/api/admin/users', { method: 'POST', body: JSON.stringify({ email, password, globalRole }) });
    if (!res.ok) return alert('Create failed');
    setEmail(''); setPassword(''); setGlobalRole('user'); load();
  }

  return <div className={styles.container}><h1>Users</h1><form onSubmit={create} className={styles.card}><div className={styles.formRow}><input className={styles.input} placeholder='Email' value={email} onChange={e=>setEmail(e.target.value)} /><input className={styles.input} type='password' placeholder='Password' value={password} onChange={e=>setPassword(e.target.value)} /><select className={styles.select} value={globalRole} onChange={e=>setGlobalRole(e.target.value)}><option value='user'>user</option><option value='superadmin'>superadmin</option></select><button className={styles.button}>Create</button></div></form><table className={styles.table}><thead><tr><th>Email</th><th>Global Role</th></tr></thead><tbody>{users.map(u=><tr key={u.id}><td>{u.email}</td><td>{u.globalRole}</td></tr>)}</tbody></table></div>;
}
