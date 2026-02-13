import Head from 'next/head';
import Link from 'next/link';
import styles from '../styles/Landing.module.css';

export default function LandingPage() {
  return (
    <>
      <Head>
        <title>Youpp | Web Siteni Hızla Yayına Al</title>
      </Head>
      <main className={styles.page}>
        <section className={styles.hero}>
          <h1>Youpp ile dakikalar içinde yayına çık</h1>
          <p>Hemen hesabını oluştur, siten otomatik oluşsun ve panelde düzenlemeye başla.</p>
          <Link href='/register' className={styles.cta}>Ücretsiz Başla</Link>
        </section>
      </main>
    </>
  );
}
