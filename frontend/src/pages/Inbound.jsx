import React, { useState, useEffect } from 'react';

export default function Inbound({ userRole, apiHeaders, triggerToast }) {
  const [purchaseOrders, setPurchaseOrders] = useState([]);
  const [products, setProducts] = useState([]);
  const [loading, setLoading] = useState(true);

  // PO Creation Form State
  const [isAddOpen, setIsAddOpen] = useState(false);
  const [supplierName, setSupplierName] = useState('');
  const [poItems, setPoItems] = useState([
    { productId: '', sku: '', quantity: 1, unitCost: 0 }
  ]);

  // Goods Receiving Modal State
  const [isReceiveOpen, setIsReceiveOpen] = useState(false);
  const [selectedPO, setSelectedPO] = useState(null);
  const [receivingQty, setReceivingQty] = useState({}); // SKU -> qty to receive in this batch

  const fetchData = async () => {
    try {
      setLoading(true);
      const [poRes, productsRes] = await Promise.all([
        fetch('http://localhost:8080/api/inbound/po', { headers: apiHeaders }),
        fetch('http://localhost:8080/api/products', { headers: apiHeaders })
      ]);

      if (poRes.ok && productsRes.ok) {
        const poData = await poRes.json();
        const productsData = await productsRes.json();
        setPurchaseOrders(poData.sort((a,b) => new Date(b.createdAt) - new Date(a.createdAt)));
        setProducts(productsData);
      }
    } catch (err) {
      console.error(err);
      triggerToast('Error loading procurement data');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchData();
  }, [userRole]);

  // Handle items rows adding in PO creation
  const handleAddItemRow = () => {
    setPoItems([...poItems, { productId: '', sku: '', quantity: 1, unitCost: 0 }]);
  };

  const handleRemoveItemRow = (index) => {
    const list = [...poItems];
    list.splice(index, 1);
    setPoItems(list);
  };

  const handleItemRowChange = (index, field, value) => {
    const list = [...poItems];
    if (field === 'productId') {
      const p = products.find(prod => prod.id === value);
      list[index].productId = value;
      list[index].sku = p ? p.sku : '';
      list[index].unitCost = p ? p.costPrice : 0;
    } else if (field === 'quantity') {
      list[index].quantity = parseFloat(value) || 0;
    } else if (field === 'unitCost') {
      list[index].unitCost = parseFloat(value) || 0;
    }
    setPoItems(list);
  };

  const handleOpenAddPO = () => {
    if (userRole !== 'Admin' && userRole !== 'SalesRep') {
      triggerToast('Access denied: Creating POs requires Admin or SalesRep roles.');
      return;
    }
    setSupplierName('');
    setPoItems([{ productId: '', sku: '', quantity: 1, unitCost: 0 }]);
    setIsAddOpen(true);
  };

  const handleCreatePOSubmit = async (e) => {
    e.preventDefault();
    if (!supplierName || poItems.some(item => !item.productId || item.quantity <= 0)) {
      alert('Supplier Name is required, and all items must have a product and a quantity > 0.');
      return;
    }

    const payload = {
      supplierName,
      items: poItems.map(item => ({
        productId: item.productId,
        sku: item.sku,
        quantity: item.quantity,
        unitCost: item.unitCost
      }))
    };

    try {
      const res = await fetch('http://localhost:8080/api/inbound/po', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          ...apiHeaders
        },
        body: JSON.stringify(payload)
      });

      if (res.ok) {
        triggerToast('Purchase Order created in Draft mode');
        setIsAddOpen(false);
        fetchData();
      } else {
        const text = await res.text();
        alert(text || 'Failed to create PO');
      }
    } catch (err) {
      console.error(err);
      alert('Error creating Purchase Order');
    }
  };

  const handleSendPO = async (id, poNum) => {
    if (userRole !== 'Admin' && userRole !== 'SalesRep') {
      triggerToast('Access denied: Sending POs to suppliers requires Admin or SalesRep roles.');
      return;
    }

    try {
      const res = await fetch(`http://localhost:8080/api/inbound/po/send/${id}`, {
        method: 'POST',
        headers: apiHeaders
      });

      if (res.ok) {
        triggerToast(`Purchase order ${poNum} marked as Sent`);
        fetchData();
      } else {
        const text = await res.text();
        alert(text || 'Failed to send PO');
      }
    } catch (err) {
      console.error(err);
      alert('Error sending PO');
    }
  };

  const handleOpenReceive = (po) => {
    if (userRole !== 'Admin' && userRole !== 'WarehouseStaff') {
      triggerToast('Access denied: Inbound shipment receiving (GRN) requires Admin or WarehouseStaff roles.');
      return;
    }

    setSelectedPO(po);
    // Initialize receiving qty inputs defaulting to quantity left to receive
    const initQty = {};
    po.items.forEach(item => {
      const left = item.quantity - item.receivedQuantity;
      initQty[item.sku] = left > 0 ? left : 0;
    });
    setReceivingQty(initQty);
    setIsReceiveOpen(true);
  };

  const handleReceiveSubmit = async (e) => {
    e.preventDefault();

    const payload = {
      receivedQty: {}
    };

    // Parse inputs to float and filter out zero values
    Object.keys(receivingQty).forEach(sku => {
      const val = parseFloat(receivingQty[sku]);
      if (val > 0) {
        payload.receivedQty[sku] = val;
      }
    });

    if (Object.keys(payload.receivedQty).length === 0) {
      alert('Please specify received quantities greater than 0.');
      return;
    }

    try {
      const res = await fetch(`http://localhost:8080/api/inbound/po/receive/${selectedPO.id}`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          ...apiHeaders
        },
        body: JSON.stringify(payload)
      });

      if (res.ok) {
        triggerToast('Goods receiving logged. Stock levels updated.');
        setIsReceiveOpen(false);
        fetchData();
      } else {
        const text = await res.text();
        alert(text || 'Failed to receive goods');
      }
    } catch (err) {
      console.error(err);
      alert('Error receiving goods');
    }
  };

  const formatPrice = (price) => {
    return new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' }).format(price);
  };

  const formatDate = (dateStr) => {
    return new Date(dateStr).toLocaleDateString('en-US', {
      month: 'short',
      day: 'numeric',
      year: 'numeric'
    });
  };

  const getStatusClass = (status) => {
    switch (status) {
      case 'Draft': return 'alert-warning';
      case 'Sent': return 'alert-warning';
      case 'Received': return 'alert-success';
      case 'Cancelled': return 'alert-danger';
      default: return 'alert-warning';
    }
  };

  return (
    <div>
      <div className="actions-panel">
        <div>
          <h2>Inbound Shipments & Procurement (PO)</h2>
          <p style={{ fontSize: '0.85rem', color: 'var(--text-muted)' }}>
            Supplier purchase orders (POs) and goods receiving notes (GRNs) logging.
          </p>
        </div>
        
        <button className="btn btn-primary" onClick={handleOpenAddPO}>
          + Create Purchase Order
        </button>
      </div>

      <div className="table-panel">
        <div className="table-wrapper">
          {loading ? (
            <div style={{ textAlign: 'center', padding: '3rem', color: 'var(--text-muted)' }}>Loading PO list...</div>
          ) : purchaseOrders.length === 0 ? (
            <div style={{ textAlign: 'center', padding: '3rem', color: 'var(--text-muted)' }}>
              No purchase orders found. Click "+ Create Purchase Order" to procure inventory.
            </div>
          ) : (
            <table className="table">
              <thead>
                <tr>
                  <th>PO Number</th>
                  <th>Supplier</th>
                  <th>Order Date</th>
                  <th>Items (Ordered / Received)</th>
                  <th>Total Amount</th>
                  <th>Status</th>
                  <th style={{ textAlign: 'right' }}>Actions</th>
                </tr>
              </thead>
              <tbody>
                {purchaseOrders.map((po) => {
                  const itemsSummary = po.items.map(item => 
                    `${item.sku}: ${item.quantity} ordered (${item.receivedQuantity} received)`
                  ).join(', ');

                  return (
                    <tr key={po.id}>
                      <td style={{ fontFamily: 'monospace', fontWeight: '600', color: 'var(--text-primary)' }}>
                        {po.poNumber}
                      </td>
                      <td>{po.supplierName}</td>
                      <td>{formatDate(po.createdAt)}</td>
                      <td style={{ fontSize: '0.85rem', maxWidth: '300px', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }} title={itemsSummary}>
                        {itemsSummary}
                      </td>
                      <td style={{ fontWeight: '600' }}>{formatPrice(po.totalAmount)}</td>
                      <td>
                        <span className={`stock-alert-pill ${getStatusClass(po.status)}`}>
                          {po.status}
                        </span>
                      </td>
                      <td style={{ textAlign: 'right' }}>
                        <div style={{ display: 'inline-flex', gap: '0.5rem' }}>
                          {po.status === 'Draft' && (
                            <button 
                              className="btn btn-outline btn-sm"
                              onClick={() => handleSendPO(po.id, po.poNumber)}
                              disabled={userRole !== 'Admin' && userRole !== 'SalesRep'}
                            >
                              Send PO
                            </button>
                          )}
                          {po.status === 'Sent' && (
                            <button 
                              className="btn btn-primary btn-sm"
                              onClick={() => handleOpenReceive(po)}
                              disabled={userRole !== 'Admin' && userRole !== 'WarehouseStaff'}
                            >
                              Receive Goods
                            </button>
                          )}
                          {(po.status === 'Received' || po.status === 'Cancelled') && (
                            <span style={{ fontSize: '0.8rem', color: 'var(--text-muted)' }}>Processed</span>
                          )}
                        </div>
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          )}
        </div>
      </div>

      {/* Create PO Modal */}
      {isAddOpen && (
        <div className="modal-overlay" onClick={() => setIsAddOpen(false)}>
          <div className="modal-content" style={{ maxWidth: '650px' }} onClick={(e) => e.stopPropagation()}>
            <div className="modal-header">
              <h3>Create Purchase Order (Draft)</h3>
              <button className="modal-close" onClick={() => setIsAddOpen(false)}>&times;</button>
            </div>
            
            <form onSubmit={handleCreatePOSubmit}>
              <div className="modal-body">
                <div className="form-group" style={{ marginBottom: '1.2rem' }}>
                  <label className="form-label">Supplier Name *</label>
                  <input 
                    type="text" 
                    className="form-input" 
                    placeholder="e.g. Samsung Semiconductor"
                    value={supplierName}
                    onChange={e => setSupplierName(e.target.value)}
                    required
                  />
                </div>

                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '0.8rem' }}>
                  <h4 style={{ fontSize: '0.95rem' }}>Order Items *</h4>
                  <button type="button" className="btn btn-outline btn-sm" onClick={handleAddItemRow}>
                    + Add Row
                  </button>
                </div>

                <div style={{ display: 'flex', flexDirection: 'column', gap: '0.8rem', maxHeight: '250px', overflowY: 'auto', paddingRight: '0.2rem' }}>
                  {poItems.map((item, index) => (
                    <div key={index} style={{ display: 'flex', gap: '0.6rem', alignItems: 'flex-end' }}>
                      <div className="form-group" style={{ flex: 2 }}>
                        <label className="form-label">Product SKU</label>
                        <select 
                          className="form-select"
                          value={item.productId}
                          onChange={e => handleItemRowChange(index, 'productId', e.target.value)}
                          required
                        >
                          <option value="">-- Choose Product --</option>
                          {products.map(p => (
                            <option key={p.id} value={p.id}>{p.sku} ({p.name})</option>
                          ))}
                        </select>
                      </div>

                      <div className="form-group" style={{ flex: 1 }}>
                        <label className="form-label">Qty</label>
                        <input 
                          type="number" 
                          className="form-input"
                          min="1"
                          value={item.quantity}
                          onChange={e => handleItemRowChange(index, 'quantity', e.target.value)}
                          required
                        />
                      </div>

                      <div className="form-group" style={{ flex: 1 }}>
                        <label className="form-label">Cost (USD)</label>
                        <input 
                          type="number" 
                          step="0.01"
                          className="form-input"
                          value={item.unitCost}
                          onChange={e => handleItemRowChange(index, 'unitCost', e.target.value)}
                          required
                        />
                      </div>

                      <button 
                        type="button" 
                        className="btn btn-danger btn-sm" 
                        style={{ height: '36px', width: '36px', padding: 0 }}
                        onClick={() => handleRemoveItemRow(index)}
                        disabled={poItems.length === 1}
                      >
                        🗑
                      </button>
                    </div>
                  ))}
                </div>
              </div>
              
              <div className="modal-footer">
                <button type="button" className="btn btn-outline" onClick={() => setIsAddOpen(false)}>
                  Cancel
                </button>
                <button type="submit" className="btn btn-primary">
                  Save Draft PO
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Receive Goods (GRN) Modal */}
      {isReceiveOpen && selectedPO && (
        <div className="modal-overlay" onClick={() => setIsReceiveOpen(false)}>
          <div className="modal-content" onClick={(e) => e.stopPropagation()}>
            <div className="modal-header">
              <h3>Goods Receiving Note (GRN): {selectedPO.poNumber}</h3>
              <button className="modal-close" onClick={() => setIsReceiveOpen(false)}>&times;</button>
            </div>
            
            <form onSubmit={handleReceiveSubmit}>
              <div className="modal-body">
                <p style={{ fontSize: '0.85rem', color: 'var(--text-muted)', marginBottom: '1.2rem' }}>
                  Log incoming quantities delivered by supplier. Increments physical inventory immediately.
                </p>
                
                <div style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
                  {selectedPO.items.map((item) => {
                    const left = item.quantity - item.receivedQuantity;
                    return (
                      <div 
                        key={item.sku} 
                        style={{ 
                          padding: '0.8rem', 
                          border: '1px solid rgba(255,255,255,0.03)', 
                          backgroundColor: 'rgba(0,0,0,0.1)', 
                          borderRadius: 'var(--radius-sm)',
                          display: 'flex',
                          justifyContent: 'space-between',
                          alignItems: 'center'
                        }}
                      >
                        <div>
                          <div style={{ fontWeight: '600' }}>{item.sku}</div>
                          <div style={{ fontSize: '0.75rem', color: 'var(--text-muted)' }}>
                            Ordered: {item.quantity} | Total Received: {item.receivedQuantity}
                          </div>
                        </div>

                        <div className="form-group" style={{ width: '120px' }}>
                          <label className="form-label">This Delivery</label>
                          <input 
                            type="number" 
                            className="form-input"
                            min="0"
                            max={left}
                            value={receivingQty[item.sku] || 0}
                            onChange={e => setReceivingQty({
                              ...receivingQty,
                              [item.sku]: parseFloat(e.target.value) || 0
                            })}
                          />
                        </div>
                      </div>
                    );
                  })}
                </div>
              </div>
              
              <div className="modal-footer">
                <button type="button" className="btn btn-outline" onClick={() => setIsReceiveOpen(false)}>
                  Cancel
                </button>
                <button type="submit" className="btn btn-primary">
                  Log Received Shipment
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
