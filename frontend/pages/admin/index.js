import Link from 'next/link';
import styles from '../../styles/Admin.module.css';

export default function AdminHome() {
  return <div className={styles.container}><div className={styles.header}><h1>Admin Dashboard</h1></div><div className={styles.card}><ul><li><Link className={styles.link} href='/admin/sites'>Sites</Link></li><li><Link className={styles.link} href='/admin/users'>Users</Link></li></ul></div></div>;
}
