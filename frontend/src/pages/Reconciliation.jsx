import React, { useState, useEffect } from 'react';
import { useT } from '../i18n.jsx';

export default function Reconciliation({ userRole, apiHeaders, triggerToast }) {
  const { t } = useT();
  const [history, setHistory] = useState([]);
  const [stockLevels, setStockLevels] = useState([]);
  const [loading, setLoading] = useState(true);
  const [view, setView] = useState('history'); // 'history' | 'new'
  const [counts, setCounts] = useState({});
  const [note, setNote] = useState('');
  const [submitting, setSubmitting] = useState(false);

  const canSubmit = userRole === 'Admin' || userRole === 'WarehouseStaff';
  const canApprove = userRole === 'Admin';

  const fetchData = async () => {
    try {
      setLoading(true);
      const [reconRes, stockRes] = await Promise.all([
        fetch('http://localhost:8080/api/reconciliation', { headers: apiHeaders }),
        fetch('http://localhost:8080/api/stock', { headers: apiHeaders }),
      ]);
      if (reconRes.ok) setHistory(await reconRes.json());
      if (stockRes.ok) {
        const stocks = await stockRes.json();
        setStockLevels(stocks);
        const initial = {};
        stocks.forEach(s => { initial[s.sku] = ''; });
        setCounts(initial);
      }
    } catch (err) {
      triggerToast('Error loading reconciliation data');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => { fetchData(); }, []);

  const handleSubmit = async () => {
    const items = stockLevels.map(s => ({
      sku: s.sku,
      actualQty: parseFloat(counts[s.sku]) || 0,
    }));

    setSubmitting(true);
    try {
      const res = await fetch('http://localhost:8080/api/reconciliation', {
        method: 'POST',
        headers: { ...apiHeaders, 'Content-Type': 'application/json' },
        body: JSON.stringify({ items, note }),
      });
      if (res.ok) {
        triggerToast(t('recon_submitted'));
        setView('history');
        setNote('');
        fetchData();
      } else {
        const err = await res.json();
        triggerToast('Error: ' + (err.error || 'Submit failed'));
      }
    } finally {
      setSubmitting(false);
    }
  };

  const handleApprove = async (id) => {
    const res = await fetch(`http://localhost:8080/api/reconciliation/${id}/approve`, {
      method: 'POST',
      headers: { ...apiHeaders, 'Content-Type': 'application/json' },
      body: JSON.stringify({ note: 'Approved' }),
    });
    if (res.ok) {
      triggerToast(t('recon_approved'));
      fetchData();
    } else {
      const err = await res.json();
      triggerToast('Error: ' + (err.error || 'Approve failed'));
    }
  };

  if (loading) return <div style={{ padding: '2rem', color: 'var(--text-muted)' }}>Loading...</div>;

  return (
    <div>
      <div style={{ display: 'flex', gap: '1rem', marginBottom: '1.5rem' }}>
        <button
          className={`btn ${view === 'history' ? 'btn-primary' : 'btn-secondary'}`}
          onClick={() => setView('history')}
        >
          {t('recon_history')}
        </button>
        {canSubmit && (
          <button
            className={`btn ${view === 'new' ? 'btn-primary' : 'btn-secondary'}`}
            onClick={() => setView('new')}
          >
            {t('recon_new')}
          </button>
        )}
      </div>

      {view === 'history' && (
        <div>
          {history.length === 0 ? (
            <p style={{ color: 'var(--text-muted)' }}>{t('recon_no_history')}</p>
          ) : (
            history.map(r => (
              <div key={r.id} className="card" style={{ marginBottom: '1rem' }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '0.75rem' }}>
                  <div>
                    <span style={{ fontWeight: 600 }}>{new Date(r.date).toLocaleDateString()}</span>
                    <span style={{ marginLeft: '1rem', color: 'var(--text-muted)', fontSize: '0.85rem' }}>
                      {t('recon_by')} {r.submittedBy}
                    </span>
                  </div>
                  <div style={{ display: 'flex', gap: '0.75rem', alignItems: 'center' }}>
                    <span className={`role-badge ${r.status === 'Approved' ? 'role-badge-admin' : 'role-badge-staff'}`}>
                      {r.status}
                    </span>
                    {r.status === 'Draft' && canApprove && (
                      <button className="btn btn-primary btn-sm" onClick={() => handleApprove(r.id)}>
                        {t('recon_approve')}
                      </button>
                    )}
                  </div>
                </div>
                <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: '0.85rem' }}>
                  <thead>
                    <tr style={{ color: 'var(--text-muted)', textAlign: 'left' }}>
                      <th style={{ padding: '0.4rem 0.6rem' }}>{t('recon_sku')}</th>
                      <th style={{ padding: '0.4rem 0.6rem' }}>{t('recon_product')}</th>
                      <th style={{ padding: '0.4rem 0.6rem', textAlign: 'right' }}>{t('recon_expected')}</th>
                      <th style={{ padding: '0.4rem 0.6rem', textAlign: 'right' }}>{t('recon_actual')}</th>
                      <th style={{ padding: '0.4rem 0.6rem', textAlign: 'right' }}>{t('recon_diff')}</th>
                    </tr>
                  </thead>
                  <tbody>
                    {(r.items || []).map(item => (
                      <tr key={item.sku} style={{ borderTop: '1px solid rgba(255,255,255,0.05)' }}>
                        <td style={{ padding: '0.4rem 0.6rem', fontFamily: 'monospace' }}>{item.sku}</td>
                        <td style={{ padding: '0.4rem 0.6rem' }}>{item.productName}</td>
                        <td style={{ padding: '0.4rem 0.6rem', textAlign: 'right' }}>{item.expectedQty}</td>
                        <td style={{ padding: '0.4rem 0.6rem', textAlign: 'right' }}>{item.actualQty}</td>
                        <td style={{
                          padding: '0.4rem 0.6rem',
                          textAlign: 'right',
                          color: item.difference === 0 ? 'var(--success)' : item.difference > 0 ? 'var(--accent-color)' : 'var(--danger)',
                          fontWeight: 600
                        }}>
                          {item.difference > 0 ? '+' : ''}{item.difference}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
                {r.note && (
                  <p style={{ marginTop: '0.5rem', fontSize: '0.8rem', color: 'var(--text-muted)' }}>
                    Note: {r.note}
                  </p>
                )}
              </div>
            ))
          )}
        </div>
      )}

      {view === 'new' && (
        <div className="card">
          <h3 style={{ marginBottom: '1rem' }}>{t('recon_title')}</h3>
          <table style={{ width: '100%', borderCollapse: 'collapse', marginBottom: '1.5rem' }}>
            <thead>
              <tr style={{ color: 'var(--text-muted)', textAlign: 'left' }}>
                <th style={{ padding: '0.5rem 0.6rem' }}>{t('recon_sku')}</th>
                <th style={{ padding: '0.5rem 0.6rem' }}>{t('recon_expected')}</th>
                <th style={{ padding: '0.5rem 0.6rem' }}>{t('recon_actual')}</th>
              </tr>
            </thead>
            <tbody>
              {stockLevels.map(s => (
                <tr key={s.sku} style={{ borderTop: '1px solid rgba(255,255,255,0.05)' }}>
                  <td style={{ padding: '0.5rem 0.6rem', fontFamily: 'monospace' }}>{s.sku}</td>
                  <td style={{ padding: '0.5rem 0.6rem', color: 'var(--text-muted)' }}>{s.stockOnHand}</td>
                  <td style={{ padding: '0.5rem 0.6rem' }}>
                    <input
                      type="number"
                      min="0"
                      className="form-input"
                      style={{ width: '100px', padding: '0.3rem 0.5rem' }}
                      placeholder={s.stockOnHand}
                      value={counts[s.sku] ?? ''}
                      onChange={e => setCounts(prev => ({ ...prev, [s.sku]: e.target.value }))}
                    />
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
          <div className="form-group" style={{ marginBottom: '1rem' }}>
            <label className="form-label">{t('recon_note')}</label>
            <input
              type="text"
              className="form-input"
              value={note}
              onChange={e => setNote(e.target.value)}
              placeholder={t('recon_note_placeholder')}
            />
          </div>
          <div style={{ display: 'flex', gap: '0.75rem' }}>
            <button className="btn btn-primary" onClick={handleSubmit} disabled={submitting}>
              {submitting ? t('recon_submitting') : t('recon_submit')}
            </button>
            <button className="btn btn-secondary" onClick={() => setView('history')}>
              {t('recon_cancel')}
            </button>
          </div>
        </div>
      )}
    </div>
  );
}
