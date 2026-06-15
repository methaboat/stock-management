import React, { useState, useEffect, useRef } from 'react';
import { useT } from '../i18n.jsx';

export default function BarcodeScanner({ userRole, apiHeaders, triggerToast }) {
  const { t } = useT();
  const [barcode, setBarcode] = useState('');
  const [status, setStatus] = useState('in');
  const [quantity, setQuantity] = useState(1);
  const [note, setNote] = useState('');
  const [result, setResult] = useState(null);
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  const [recentScans, setRecentScans] = useState([]);
  const inputRef = useRef(null);

  // Auto-focus the barcode input so USB scanner can type directly into it
  useEffect(() => {
    inputRef.current?.focus();
  }, [result]);

  const handleScan = async (e) => {
    e.preventDefault();
    if (!barcode.trim()) return;

    setError('');
    setResult(null);
    setLoading(true);

    try {
      const res = await fetch('http://localhost:8080/api/stock/scan', {
        method: 'POST',
        headers: { ...apiHeaders, 'Content-Type': 'application/json' },
        body: JSON.stringify({
          barcode: barcode.trim(),
          status,
          quantity: Number(quantity),
          note,
        }),
      });

      const data = await res.json();

      if (res.ok) {
        setResult(data);
        setRecentScans(prev => [{
          barcode: barcode.trim(),
          productName: data.product.name,
          sku: data.product.sku,
          status,
          quantity: Number(quantity),
          newStock: data.stock.stockOnHand,
          time: new Date().toLocaleTimeString(),
        }, ...prev.slice(0, 9)]);
        setBarcode('');
        setQuantity(1);
        setNote('');
      } else {
        setError(data.error || 'Scan failed');
      }
    } catch (err) {
      setError('Cannot connect to backend');
    } finally {
      setLoading(false);
    }
  };

  const handleBarcodeKeyDown = (e) => {
    // USB scanners typically send Enter after the barcode
    if (e.key === 'Enter') {
      handleScan(e);
    }
  };

  return (
    <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1.5rem' }}>
      {/* Scan Panel */}
      <div>
        <div className="card" style={{ marginBottom: '1rem' }}>
          <h3 style={{ marginBottom: '1.25rem' }}>{t('scanner_title')}</h3>

          {/* Status Toggle */}
          <div style={{ display: 'flex', gap: '0.5rem', marginBottom: '1.25rem' }}>
            <button
              className={`btn ${status === 'in' ? 'btn-primary' : 'btn-secondary'}`}
              style={{ flex: 1 }}
              onClick={() => setStatus('in')}
              type="button"
            >
              {t('scanner_in')}
            </button>
            <button
              className={`btn ${status === 'out' ? 'btn-primary' : 'btn-secondary'}`}
              style={{ flex: 1, ...(status === 'out' ? { backgroundColor: 'var(--danger)', borderColor: 'var(--danger)' } : {}) }}
              onClick={() => setStatus('out')}
              type="button"
            >
              {t('scanner_out')}
            </button>
          </div>

          <form onSubmit={handleScan}>
            <div className="form-group" style={{ marginBottom: '1rem' }}>
              <label className="form-label">{t('scanner_barcode_label')}</label>
              <input
                ref={inputRef}
                type="text"
                className="form-input"
                placeholder={t('scanner_barcode_placeholder')}
                value={barcode}
                onChange={e => setBarcode(e.target.value)}
                onKeyDown={handleBarcodeKeyDown}
                autoComplete="off"
                style={{ fontSize: '1.1rem', letterSpacing: '0.05em' }}
              />
            </div>

            <div className="form-group" style={{ marginBottom: '1rem' }}>
              <label className="form-label">{t('scanner_qty')}</label>
              <input
                type="number"
                className="form-input"
                min="1"
                value={quantity}
                onChange={e => setQuantity(e.target.value)}
              />
            </div>

            <div className="form-group" style={{ marginBottom: '1.25rem' }}>
              <label className="form-label">{t('scanner_note')}</label>
              <input
                type="text"
                className="form-input"
                placeholder={t('scanner_note_placeholder')}
                value={note}
                onChange={e => setNote(e.target.value)}
              />
            </div>

            <button
              type="submit"
              className="btn btn-primary"
              style={{ width: '100%' }}
              disabled={loading || !barcode.trim()}
            >
              {loading ? t('scanner_processing') : `${t('scanner_confirm')} ${status.toUpperCase()}`}
            </button>
          </form>
        </div>

        {/* Error */}
        {error && (
          <div style={{
            padding: '0.75rem 1rem',
            borderRadius: '6px',
            backgroundColor: 'rgba(239,68,68,0.1)',
            border: '1px solid rgba(239,68,68,0.3)',
            color: 'var(--danger)',
            marginBottom: '1rem',
          }}>
            ⚠ {error}
          </div>
        )}

        {/* Result Card */}
        {result && (
          <div style={{
            padding: '1rem',
            borderRadius: '8px',
            backgroundColor: 'rgba(34,197,94,0.08)',
            border: '1px solid rgba(34,197,94,0.25)',
          }}>
            <div style={{ color: 'var(--success)', fontWeight: 700, marginBottom: '0.5rem', fontSize: '1rem' }}>
              ✓ {result.quantityChanged > 0 ? '+' : ''}{result.quantityChanged} units recorded
            </div>
            <div style={{ fontSize: '0.9rem', lineHeight: '1.8' }}>
              <div><span style={{ color: 'var(--text-muted)' }}>{t('scanner_result_product')}:</span> {result.product.name}</div>
              <div><span style={{ color: 'var(--text-muted)' }}>{t('scanner_result_sku')}:</span> {result.product.sku}</div>
              <div><span style={{ color: 'var(--text-muted)' }}>{t('scanner_result_stock')}:</span> <strong>{result.stock.stockOnHand}</strong></div>
            </div>
          </div>
        )}
      </div>

      {/* Recent Scans */}
      <div className="card">
        <h3 style={{ marginBottom: '1rem' }}>{t('scanner_recent')}</h3>
        {recentScans.length === 0 ? (
          <p style={{ color: 'var(--text-muted)', fontSize: '0.9rem' }}>{t('scanner_no_scans')}</p>
        ) : (
          <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: '0.85rem' }}>
            <thead>
              <tr style={{ color: 'var(--text-muted)', textAlign: 'left' }}>
                <th style={{ padding: '0.4rem 0.5rem' }}>{t('scanner_time')}</th>
                <th style={{ padding: '0.4rem 0.5rem' }}>{t('scanner_product')}</th>
                <th style={{ padding: '0.4rem 0.5rem', textAlign: 'center' }}>{t('scanner_status')}</th>
                <th style={{ padding: '0.4rem 0.5rem', textAlign: 'right' }}>{t('scanner_qty_col')}</th>
                <th style={{ padding: '0.4rem 0.5rem', textAlign: 'right' }}>{t('scanner_stock')}</th>
              </tr>
            </thead>
            <tbody>
              {recentScans.map((s, i) => (
                <tr key={i} style={{ borderTop: '1px solid rgba(255,255,255,0.05)' }}>
                  <td style={{ padding: '0.4rem 0.5rem', color: 'var(--text-muted)', fontSize: '0.8rem' }}>{s.time}</td>
                  <td style={{ padding: '0.4rem 0.5rem' }}>
                    <div style={{ fontWeight: 500 }}>{s.productName}</div>
                    <div style={{ color: 'var(--text-muted)', fontSize: '0.75rem', fontFamily: 'monospace' }}>{s.sku}</div>
                  </td>
                  <td style={{ padding: '0.4rem 0.5rem', textAlign: 'center' }}>
                    <span style={{
                      padding: '0.15rem 0.5rem',
                      borderRadius: '4px',
                      fontSize: '0.75rem',
                      fontWeight: 600,
                      backgroundColor: s.status === 'in' ? 'rgba(34,197,94,0.15)' : 'rgba(239,68,68,0.15)',
                      color: s.status === 'in' ? 'var(--success)' : 'var(--danger)',
                    }}>
                      {s.status.toUpperCase()}
                    </span>
                  </td>
                  <td style={{ padding: '0.4rem 0.5rem', textAlign: 'right' }}>{s.quantity}</td>
                  <td style={{ padding: '0.4rem 0.5rem', textAlign: 'right', fontWeight: 600 }}>{s.newStock}</td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>
    </div>
  );
}
