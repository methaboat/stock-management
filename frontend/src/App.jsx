import React, { useState, useEffect } from 'react';
import { useT } from './i18n.jsx';
import Dashboard from './pages/Dashboard';
import Catalog from './pages/Catalog';
import StockLevels from './pages/StockLevels';
import Inbound from './pages/Inbound';
import Outbound from './pages/Outbound';
import Analytics from './pages/Analytics';
import Reconciliation from './pages/Reconciliation';
import BarcodeScanner from './pages/BarcodeScanner';

export default function App() {
  const { lang, setLang, t } = useT();

  // Navigation & Authentication
  const [currentView, setCurrentView] = useState('dashboard');
  const [token, setToken] = useState(localStorage.getItem('token') || '');
  const [username, setUsername] = useState(localStorage.getItem('username') || '');
  const [role, setRole] = useState(localStorage.getItem('role') || '');

  // Login Form
  const [loginUser, setLoginUser] = useState('');
  const [loginPass, setLoginPass] = useState('');
  const [loginError, setLoginError] = useState('');
  const [isLoggingIn, setIsLoggingIn] = useState(false);

  // UI styling states
  const [theme, setTheme] = useState('dark');
  const [toast, setToast] = useState({ message: '', visible: false });

  // Update session
  const setSession = (session) => {
    if (session) {
      setToken(session.token);
      setUsername(session.username);
      setRole(session.role);
      localStorage.setItem('token', session.token);
      localStorage.setItem('username', session.username);
      localStorage.setItem('role', session.role);
    } else {
      setToken('');
      setUsername('');
      setRole('');
      localStorage.removeItem('token');
      localStorage.removeItem('username');
      localStorage.removeItem('role');
    }
  };

  // Toast trigger
  const triggerToast = (message) => {
    setToast({ message, visible: true });
    setTimeout(() => {
      setToast({ message: '', visible: false });
    }, 4000);
  };

  // API auth headers helper
  const apiHeaders = {
    'Authorization': `Bearer ${token}`
  };

  const handleLoginSubmit = async (e) => {
    e.preventDefault();
    if (!loginUser || !loginPass) {
      setLoginError('Username and Password are required.');
      return;
    }

    setLoginError('');
    setIsLoggingIn(true);
    try {
      const res = await fetch('http://localhost:8080/api/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username: loginUser, password: loginPass })
      });

      if (res.ok) {
        const data = await res.json();
        setSession(data);
        triggerToast(`${t('login_welcome')}, ${data.username}!`);
      } else {
        const text = await res.text();
        setLoginError(text || 'Invalid username or password.');
      }
    } catch (err) {
      console.error(err);
      setLoginError('Could not connect to Go backend. Verify it is running on port 8080.');
    } finally {
      setIsLoggingIn(false);
    }
  };

  const handleLogout = () => {
    setSession(null);
    triggerToast('Logged out successfully');
  };

  const toggleTheme = () => {
    const nextTheme = theme === 'dark' ? 'light' : 'dark';
    setTheme(nextTheme);
    if (nextTheme === 'light') {
      document.body.classList.add('light-theme');
    } else {
      document.body.classList.remove('light-theme');
    }
  };

  // Return login card if not logged in
  if (!token) {
    return (
      <div className="login-container">
        <div className="login-card">
          <div style={{ textAlign: 'center', marginBottom: '2rem' }}>
            <h1 style={{ fontSize: '2rem', marginBottom: '0.5rem' }}>
              Stock<span style={{ color: 'var(--accent-color)' }}>Flow</span>
            </h1>
            <p style={{ color: 'var(--text-muted)', fontSize: '0.9rem' }}>
              {t('login_subtitle')}
            </p>
          </div>

          {loginError && (
            <div style={{ 
              color: 'var(--danger)', 
              fontSize: '0.85rem', 
              padding: '0.6rem 0.8rem', 
              borderRadius: '4px', 
              backgroundColor: 'rgba(239,68,68,0.08)',
              border: '1px solid rgba(239,68,68,0.15)',
              marginBottom: '1.2rem'
            }}>
              ⚠️ {loginError}
            </div>
          )}

          <form onSubmit={handleLoginSubmit}>
            <div className="form-group" style={{ marginBottom: '1rem' }}>
              <label className="form-label">{t('login_username')}</label>
              <input
                type="text"
                className="form-input"
                placeholder="e.g. admin"
                value={loginUser}
                onChange={e => setLoginUser(e.target.value)}
                required
              />
            </div>

            <div className="form-group" style={{ marginBottom: '1.5rem' }}>
              <label className="form-label">{t('login_password')}</label>
              <input
                type="password"
                className="form-input"
                placeholder="••••••••"
                value={loginPass}
                onChange={e => setLoginPass(e.target.value)}
                required
              />
            </div>

            <button type="submit" className="btn btn-primary" style={{ width: '100%' }} disabled={isLoggingIn}>
              {isLoggingIn ? t('login_signing_in') : t('login_submit')}
            </button>
          </form>

          {/* Seeding guide helper */}
          <div style={{ 
            marginTop: '2rem', 
            padding: '1rem', 
            backgroundColor: 'rgba(255,255,255,0.02)', 
            border: '1px solid rgba(255,255,255,0.04)',
            borderRadius: 'var(--radius-sm)',
            fontSize: '0.8rem'
          }}>
            <div style={{ fontWeight: '600', color: 'var(--text-secondary)', marginBottom: '0.4rem' }}>{t('login_demo')}:</div>
            <div style={{ color: 'var(--text-muted)' }}>
              • **Admin**: <code>admin</code> / <code>adminpass</code><br />
              • **Staff**: <code>staff</code> / <code>staffpass</code><br />
              • **Sales**: <code>sales</code> / <code>salespass</code>
            </div>
          </div>
        </div>
      </div>
    );
  }

  // Render main layout
  return (
    <div className="app-layout">
      {/* Sidebar Navigation */}
      <aside className="sidebar">
        <div className="sidebar-brand">
          <span>Stock</span><span className="logo-highlight">Flow</span>
          <span className="brand-dot">.</span>
        </div>

        <nav className="sidebar-nav">
          <div 
            className={`nav-item ${currentView === 'dashboard' ? 'active' : ''}`}
            onClick={() => setCurrentView('dashboard')}
          >
            📊 {t('nav_dashboard')}
          </div>
          <div
            className={`nav-item ${currentView === 'catalog' ? 'active' : ''}`}
            onClick={() => setCurrentView('catalog')}
          >
            📋 {t('nav_catalog')}
          </div>
          <div
            className={`nav-item ${currentView === 'stock' ? 'active' : ''}`}
            onClick={() => setCurrentView('stock')}
          >
            📦 {t('nav_stock')}
          </div>
          <div
            className={`nav-item ${currentView === 'inbound' ? 'active' : ''}`}
            onClick={() => setCurrentView('inbound')}
          >
            📥 {t('nav_inbound')}
          </div>
          <div
            className={`nav-item ${currentView === 'outbound' ? 'active' : ''}`}
            onClick={() => setCurrentView('outbound')}
          >
            📤 {t('nav_outbound')}
          </div>
          <div
            className={`nav-item ${currentView === 'analytics' ? 'active' : ''}`}
            onClick={() => setCurrentView('analytics')}
          >
            📈 {t('nav_analytics')}
          </div>
          <div
            className={`nav-item ${currentView === 'scanner' ? 'active' : ''}`}
            onClick={() => setCurrentView('scanner')}
          >
            📷 {t('nav_scanner')}
          </div>
          <div
            className={`nav-item ${currentView === 'reconciliation' ? 'active' : ''}`}
            onClick={() => setCurrentView('reconciliation')}
          >
            ✅ {t('nav_reconciliation')}
          </div>
        </nav>

        {/* Sidebar Footer User Card */}
        <div className="sidebar-footer">
          <div style={{ fontWeight: '600', fontSize: '0.9rem', color: 'var(--text-primary)', marginBottom: '0.2rem' }}>
            👤 {username}
          </div>
          <span className={`role-badge role-badge-${role === 'Admin' ? 'admin' : role === 'WarehouseStaff' ? 'staff' : 'sales'}`}>
            {role === 'WarehouseStaff' ? 'Staff' : role === 'SalesRep' ? 'Sales' : role}
          </span>
        </div>
      </aside>

      {/* Main View Panel */}
      <div className="main-view">
        <header className="topbar">
          <h2 style={{ fontSize: '1.25rem', textTransform: 'capitalize' }}>
            {currentView === 'inbound' ? 'Inbound Logistics' : currentView === 'outbound' ? 'Outbound Logistics' : currentView}
          </h2>

          <div style={{ display: 'flex', gap: '0.8rem', alignItems: 'center' }}>
            <button
              className="btn btn-secondary btn-sm"
              onClick={() => setLang(lang === 'th' ? 'en' : 'th')}
              title="Switch language / เปลี่ยนภาษา"
              style={{ fontWeight: 600, minWidth: '2.5rem' }}
            >
              {lang === 'th' ? 'EN' : 'TH'}
            </button>
            <button
              className="theme-toggle-btn"
              onClick={toggleTheme}
              title={theme === 'dark' ? 'Switch to Light Mode' : 'Switch to Dark Mode'}
            >
              {theme === 'dark' ? '☀️' : '🌙'}
            </button>
            <button className="btn btn-secondary btn-sm" onClick={handleLogout}>
              {t('topbar_logout')}
            </button>
          </div>
        </header>

        <div className="view-container">
          {currentView === 'dashboard' && (
            <Dashboard 
              userRole={role} 
              setSession={setSession} 
              triggerToast={triggerToast} 
              apiHeaders={apiHeaders} 
            />
          )}
          {currentView === 'catalog' && (
            <Catalog 
              userRole={role} 
              apiHeaders={apiHeaders} 
              triggerToast={triggerToast} 
            />
          )}
          {currentView === 'stock' && (
            <StockLevels 
              userRole={role} 
              apiHeaders={apiHeaders} 
              triggerToast={triggerToast} 
            />
          )}
          {currentView === 'inbound' && (
            <Inbound 
              userRole={role} 
              apiHeaders={apiHeaders} 
              triggerToast={triggerToast} 
            />
          )}
          {currentView === 'outbound' && (
            <Outbound 
              userRole={role} 
              apiHeaders={apiHeaders} 
              triggerToast={triggerToast} 
            />
          )}
          {currentView === 'analytics' && (
            <Analytics
              userRole={role}
              apiHeaders={apiHeaders}
              triggerToast={triggerToast}
            />
          )}
          {currentView === 'scanner' && (
            <BarcodeScanner
              userRole={role}
              apiHeaders={apiHeaders}
              triggerToast={triggerToast}
            />
          )}
          {currentView === 'reconciliation' && (
            <Reconciliation
              userRole={role}
              apiHeaders={apiHeaders}
              triggerToast={triggerToast}
            />
          )}
        </div>
      </div>

      {/* Floating alert notifications */}
      {toast.visible && (
        <div className="toast-alert">
          <span style={{ color: 'var(--success)' }}>✓</span>
          <span>{toast.message}</span>
        </div>
      )}
    </div>
  );
}
