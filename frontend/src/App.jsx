import { Routes, Route, Navigate, useNavigate } from 'react-router-dom';
import IncidentList from './pages/IncidentList.jsx';
import IncidentDetail from './pages/IncidentDetail.jsx';
import { AuthProvider } from './AuthContext.jsx';

function Brand() {
  const navigate = useNavigate();
  return (
    <button className="wr-brand" onClick={() => navigate('/')}>
      <span className="wr-brand-mark">
        <span className="wr-brand-dot" />
      </span>
      <span className="wr-brand-name">War Room</span>
    </button>
  );
}

export default function App() {
  return (
    <AuthProvider>
      <div className="wr-app">
        <div className="wr-topbar">
          <Brand />
        </div>
        <div className="wr-main">
          <Routes>
            <Route path="/" element={<IncidentList />} />
            <Route path="/incidents/:id" element={<IncidentDetail />} />
            <Route path="/incidents/:id/:tab" element={<IncidentDetail />} />
            {/* The bot link lands on /dashboard?token=…; the token is captured
                by AuthProvider, so just send the user to the incident list. */}
            <Route path="/dashboard" element={<Navigate to="/" replace />} />
          </Routes>
        </div>
      </div>
    </AuthProvider>
  );
}
