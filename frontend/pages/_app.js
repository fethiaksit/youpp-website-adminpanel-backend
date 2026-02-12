import Head from 'next/head';
import '../styles/globals.css';

export default function App({ Component, pageProps }) {
  return (
    <>
      <Head>
        <meta name='robots' content='index,follow' />
        <meta name='googlebot' content='index,follow' />
      </Head>
      <Component {...pageProps} />
    </>
  );
}
