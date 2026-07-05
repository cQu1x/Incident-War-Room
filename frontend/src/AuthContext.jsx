import { createContext, useContext, useEffect, useState } from 'react';
import { captureTokenFromURL, getToken, clearToken } from './auth.js';
import { verifyToken } from './api.js';

const AuthContext = createContext(null);

export function useAuth() {
  return useContext(AuthContext);
}

// AuthProvider captures the dashboard token from the /dashboard link, verifies
// it against the incident-service, and only renders the app once a valid token
// is present. The verified chat id is exposed so pages can scope their queries.
export function AuthProvider({ children }) {
  const [state, setState] = useState({ status: 'checking', chatId: null });

  useEffect(() => {
    captureTokenFromURL();

    if (!getToken()) {
      setState({ status: 'unauthed', chatId: null });
      return;
    }

    let cancelled = false;
    verifyToken()
      .then((claims) => {
        if (!cancelled) setState({ status: 'authed', chatId: claims.chatId });
      })
      .catch(() => {
        clearToken();
        if (!cancelled) setState({ status: 'unauthed', chatId: null });
      });

    return () => {
      cancelled = true;
    };
  }, []);

  if (state.status === 'checking') {
    return <AuthScreen>Checking your dashboard link…</AuthScreen>;
  }

  if (state.status === 'unauthed') {
    return (
      <AuthScreen>
        This dashboard is private. Open it from the Telegram bot with the{' '}
        <code>/dashboard</code> command — the link it sends carries your personal
        access token.
      </AuthScreen>
    );
  }

  return <AuthContext.Provider value={{ chatId: state.chatId }}>{children}</AuthContext.Provider>;
}

function AuthScreen({ children }) {
  return (
    <div
      style={{
        minHeight: '100vh',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        padding: 24,
      }}
    >
      <div
        style={{
          maxWidth: 420,
          textAlign: 'center',
          fontSize: 14,
          lineHeight: 1.6,
          color: '#4b4a46',
        }}
      >
        {children}
      </div>
    </div>
  );
}
