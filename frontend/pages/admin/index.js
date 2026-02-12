import Head from 'next/head';
import Link from 'next/link';
import styles from '../../styles/Admin.module.css';

export default function AdminHome() {
  return (
    <>
      <Head>
        <title>Super Admin Panel | Youpp</title>
        <meta
          name='description'
          content='Youpp Super Admin panel landing page. Sign in to manage sites, users, and platform operations.'
        />
        <meta name='robots' content='index,follow' />
      </Head>
      <div className={styles.container}>
        <div className={styles.header}>
          <h1>Super Admin Panel</h1>
          <p>
            This is the public landing page for the Youpp Super Admin panel. Search engines can index this page.
            Sign in to access protected admin actions.
          </p>
        </div>
        <div className={styles.card}>
          <ul>
            <li>
              <Link className={styles.link} href='/login'>
                Go to Login
              </Link>
            </li>
          </ul>
        </div>
      </div>
    </>
  );
}
