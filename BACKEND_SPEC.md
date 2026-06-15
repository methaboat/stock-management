# Backend Specification — StockFlow

## Stack

| Layer | Technology |
|-------|-----------|
| Language | Go 1.25+ |
| Framework | Gin v1.9.1 |
| Database | MongoDB (via mongo-driver v1.12) |
| Auth | JWT (HS256, `golang-jwt/jwt/v5`) |
| Password | bcrypt |

Base URL: `http://localhost:8080`

---

## Authentication

All endpoints except `/api/login` and `/health/*` require a JWT Bearer token.

**Header:**
```
Authorization: Bearer <token>
```

Token payload:
```json
{ "username": "admin", "role": "Admin", "exp": 1234567890 }
```

Token TTL: **24 hours**

### Roles & Permissions

| Role | Description |
|------|-------------|
| `Admin` | Full access to all endpoints |
| `WarehouseStaff` | Stock adjustments, scan, receive PO, ship SO, create RMA, submit reconciliation |
| `SalesRep` | Create products, create PO, send PO, create SO |

Default seeded users:

| Username | Password | Role |
|----------|----------|------|
| admin | adminpass | Admin |
| staff | staffpass | WarehouseStaff |
| sales | salespass | SalesRep |

---

## MongoDB Collections & Schemas

### `users`
```json
{
  "_id":          ObjectId,
  "username":     string,       // unique
  "passwordHash": string,       // bcrypt hash
  "role":         string        // "Admin" | "WarehouseStaff" | "SalesRep"
}
```

### `products`
```json
{
  "_id":          ObjectId,
  "sku":          string,       // unique
  "name":         string,
  "barcode":      string,
  "category":     string,
  "brand":        string,
  "costPrice":    float64,
  "sellingPrice": float64,
  "supplierIds":  [string],
  "createdAt":    ISODate,
  "updatedAt":    ISODate
}
```
Indexes: `sku` (unique)

### `stocks`
```json
{
  "_id":            ObjectId,
  "productId":      ObjectId,   // ref → products._id
  "sku":            string,
  "stockOnHand":    float64,    // physical qty on shelf
  "committedStock": float64,    // reserved for pending SOs
  "availableStock": float64,    // stockOnHand - committedStock
  "reorderPoint":   float64,    // triggers low-stock alert when availableStock ≤ reorderPoint
  "location": {
    "warehouse":    string,
    "aisle":        string,
    "bin":          string
  }
}
```
Indexes: `sku` (unique), `productId`

Stock record is **auto-created** (all zeros) when a product is created via `POST /api/products`.

### `purchaseorders`
```json
{
  "_id":          ObjectId,
  "poNumber":     string,       // auto-generated "PO-<timestamp>"
  "supplierName": string,
  "status":       string,       // "Draft" → "Sent" → "Received"
  "items": [{
    "productId":        ObjectId,
    "sku":              string,
    "quantity":         float64,
    "unitCost":         float64,
    "receivedQuantity": float64
  }],
  "totalAmount":  float64,
  "createdAt":    ISODate,
  "updatedAt":    ISODate
}
```

PO status transitions:
```
Draft → Sent (POST /inbound/po/send/:id)
Sent  → Received (POST /inbound/po/receive/:id, when all items fully received)
```

### `salesorders`
```json
{
  "_id":          ObjectId,
  "soNumber":     string,       // auto-generated "SO-<timestamp>"
  "customerName": string,
  "status":       string,       // "Pending" → "Shipped"
  "items": [{
    "productId":       ObjectId,
    "sku":             string,
    "quantity":        float64,
    "unitPrice":       float64,
    "shippedQuantity": float64
  }],
  "totalAmount":  float64,
  "createdAt":    ISODate,
  "updatedAt":    ISODate
}
```

SO status transitions:
```
Pending → Shipped (POST /outbound/so/ship/:id)
```

On SO create: `availableStock -= qty`, `committedStock += qty`
On SO ship: `stockOnHand -= qty`, `committedStock -= qty`

### `rmas`
```json
{
  "_id":          ObjectId,
  "salesOrderId": ObjectId,     // ref → salesorders._id
  "productId":    ObjectId,     // ref → products._id
  "sku":          string,
  "quantity":     float64,
  "reason":       string,
  "status":       string,       // "Restocked" | "Damaged"
  "createdAt":    ISODate
}
```

If `status = "Restocked"`: `stockOnHand += qty`, `availableStock += qty`
If `status = "Damaged"`: no stock change

### `auditlogs`
```json
{
  "_id":        ObjectId,
  "username":   string,
  "action":     string,     // "Create" | "Update" | "Adjust" | "Receive" | "Reserve" | "Ship" | "Return" | "Send" | "Delete"
  "entityType": string,     // "Product" | "Stock" | "PO" | "SO"
  "entityId":   string,
  "sku":        string,
  "qtyChange":  float64,
  "note":       string,
  "createdAt":  ISODate
}
```

### `reconciliations`
```json
{
  "_id":         ObjectId,
  "date":        ISODate,
  "status":      string,        // "Draft" | "Approved"
  "items": [{
    "productId":   ObjectId,
    "sku":         string,
    "productName": string,
    "expectedQty": float64,     // from stocks.stockOnHand
    "actualQty":   float64,     // physical count entered by staff
    "difference":  float64      // actualQty - expectedQty
  }],
  "submittedBy": string,
  "approvedBy":  string,        // omitempty
  "approvedAt":  ISODate,       // omitempty
  "note":        string,
  "createdAt":   ISODate
}
```

---

## API Endpoints

### Health

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/health/liveness` | None | Returns `{"status":"UP"}` |
| GET | `/health/readiness` | None | Returns `{"status":"READY"}` |

---

### Auth

#### `POST /api/login`

Request:
```json
{ "username": "admin", "password": "adminpass" }
```

Response `200`:
```json
{
  "token":    "eyJ...",
  "username": "admin",
  "role":     "Admin"
}
```

Response `401`:
```json
{ "error": "invalid credentials" }
```

---

### Products

#### `GET /api/products`
Roles: All authenticated

Response `200`: array of Product objects
```json
[{
  "id":           "64abc...",
  "sku":          "SKU-001",
  "name":         "Wireless Mouse",
  "barcode":      "8850001000011",
  "category":     "Electronics",
  "brand":        "Logitech",
  "costPrice":    350,
  "sellingPrice": 590,
  "supplierIds":  ["SUP-001"],
  "createdAt":    "2026-06-15T10:00:00Z",
  "updatedAt":    "2026-06-15T10:00:00Z"
}]
```

#### `POST /api/products`
Roles: `Admin`, `SalesRep`

Request:
```json
{
  "sku":          "SKU-001",      // required, must be unique
  "name":         "Wireless Mouse", // required
  "barcode":      "8850001000011",
  "category":     "Electronics",
  "brand":        "Logitech",
  "costPrice":    350,            // required, > 0
  "sellingPrice": 590,            // required, > 0
  "supplierIds":  ["SUP-001"]
}
```

Response `200`: created Product object
Side effect: creates a zero-quantity stock record for this product

#### `PUT /api/products/:id`
Roles: `Admin`

Request: same fields as POST (except `sku` cannot be changed)

Response `200`: `{}`

#### `DELETE /api/products/:id`
Roles: `Admin`

Response `200`: `{}`
Side effect: deletes the associated stock record

---

### Stock

#### `GET /api/stock`
Roles: All authenticated

Response `200`: array of StockLevel objects
```json
[{
  "id":             "64abc...",
  "productId":      "64abc...",
  "sku":            "SKU-001",
  "stockOnHand":    85,
  "committedStock": 5,
  "availableStock": 80,
  "reorderPoint":   20,
  "location": {
    "warehouse": "WH-Main",
    "aisle":     "A",
    "bin":       "A-01"
  }
}]
```

#### `POST /api/stock/adjust`
Roles: `Admin`, `WarehouseStaff`

Updates an existing stock record (sets absolute values, not incremental).

Request:
```json
{
  "stockId":      "64abc...",   // required — ID of existing StockLevel
  "stockOnHand":  100,          // required, ≥ 0
  "reorderPoint": 20,           // required
  "location": {
    "warehouse": "WH-Main",
    "aisle":     "A",
    "bin":       "A-01"
  },
  "note":         "Cycle count correction"  // required
}
```

Response `200`: `{}`

#### `GET /api/stock/low-alerts`
Roles: All authenticated

Returns stock records where `availableStock ≤ reorderPoint`

Response `200`: array of StockLevel objects (same shape as `/api/stock`)

#### `POST /api/stock/scan`
Roles: `Admin`, `WarehouseStaff`

Looks up product by barcode and adjusts stock in or out.

Request:
```json
{
  "barcode":  "8850001000011",  // required
  "status":   "in",            // required: "in" | "out"
  "quantity": 5,               // required, > 0
  "note":     "Received from dock B"
}
```

Response `200`:
```json
{
  "product":         { /* Product object */ },
  "stock":           { /* StockLevel object */ },
  "quantityChanged": 5
}
```

---

### Inbound (Purchase Orders)

#### `GET /api/inbound/po`
Roles: All authenticated

Response `200`: array of PurchaseOrder objects

#### `POST /api/inbound/po`
Roles: `Admin`, `SalesRep`

`poNumber` and `totalAmount` are auto-generated.

Request:
```json
{
  "supplierName": "TechSource Co.",   // required
  "items": [{
    "productId": "64abc...",          // required
    "sku":       "SKU-001",           // required
    "quantity":  100,                 // required
    "unitCost":  350                  // required
  }]
}
```

Response `200`: created PurchaseOrder object (status = `"Draft"`)

#### `POST /api/inbound/po/send/:id`
Roles: `Admin`, `SalesRep`

Transitions PO from `Draft` → `Sent`. No request body needed.

Response `200`: `{}`
Error if PO is not in `Draft` status.

#### `POST /api/inbound/po/receive/:id`
Roles: `Admin`, `WarehouseStaff`

Records goods received. Can do partial receives. When all items are fully received the PO moves to `Received`.

Request:
```json
{
  "receivedQty": {
    "SKU-001": 50,
    "SKU-002": 25
  }
}
```

Response `200`: `{}`
Side effect: increments `stockOnHand` and `availableStock` for each received SKU.
Error if PO is not in `Sent` status.

---

### Outbound (Sales Orders & Returns)

#### `GET /api/outbound/so`
Roles: All authenticated

Response `200`: array of SalesOrder objects

#### `POST /api/outbound/so`
Roles: `Admin`, `SalesRep`

Validates available stock before creating. Reserves stock immediately on create.

Request:
```json
{
  "customerName": "ABC Company",   // required
  "items": [{
    "productId": "64abc...",       // required
    "sku":       "SKU-001",        // required
    "quantity":  5,                // required
    "unitPrice": 590               // required
  }]
}
```

Response `200`: created SalesOrder object (status = `"Pending"`)
Error `400` if any item has insufficient `availableStock`.

#### `POST /api/outbound/so/ship/:id`
Roles: `Admin`, `WarehouseStaff`

Ships all items in a sales order. Deducts `stockOnHand` and `committedStock`.

Response `200`: `{}`
Error if SO is not in `Pending` or `Processing` status.

#### `GET /api/outbound/rma`
Roles: All authenticated

Response `200`: array of ReturnsRMA objects

#### `POST /api/outbound/rma`
Roles: `Admin`, `WarehouseStaff`

Request:
```json
{
  "salesOrderId": "64abc...",             // required
  "productId":    "64abc...",             // required
  "sku":          "SKU-001",              // required
  "quantity":     1,                      // required, > 0
  "reason":       "Defective unit",
  "status":       "Restocked"             // required: "Restocked" | "Damaged"
}
```

Response `200`: created ReturnsRMA object

---

### Audit Logs

#### `GET /api/audit`
Roles: All authenticated

Response `200`: array of AuditLog objects
```json
[{
  "id":         "64abc...",
  "username":   "admin",
  "action":     "Receive",
  "entityType": "Stock",
  "entityId":   "64abc...",
  "sku":        "SKU-001",
  "qtyChange":  50,
  "note":       "Received shipment for PO PO-2024-001",
  "createdAt":  "2026-06-15T10:00:00Z"
}]
```

---

### Reports

#### `GET /api/reports/valuation`
Roles: All authenticated

Returns inventory valuation using FIFO, LIFO, and Average Cost methods. Calculations are based on PO cost history.

Response `200`:
```json
{
  "items": [{
    "sku":               "SKU-001",
    "productName":       "Wireless Mouse",
    "stockOnHand":       85,
    "baseCost":          350,
    "fifoValuation":     29750,
    "lifoValuation":     29750,
    "avgCostValuation":  29750
  }],
  "totalFifo":        125000,
  "totalLifo":        124000,
  "totalAverageCost": 124500
}
```

---

### Reconciliation

#### `GET /api/reconciliation`
Roles: All authenticated

Response `200`: array of Reconciliation objects

#### `POST /api/reconciliation`
Roles: `Admin`, `WarehouseStaff`

Submits a physical stock count. System compares actual counts against `stockOnHand` and records differences.

Request:
```json
{
  "items": [{
    "sku":       "SKU-001",   // required
    "actualQty": 83           // required, ≥ 0
  }],
  "note": "End of month cycle count"
}
```

Response `200`: created Reconciliation object (status = `"Draft"`)

#### `POST /api/reconciliation/:id/approve`
Roles: `Admin`

Approves a draft reconciliation. Updates stock `stockOnHand` and `availableStock` to the actual counts.

Request (optional):
```json
{ "note": "Approved after physical verification" }
```

Response `200`: `{}`

---

## Error Response Format

All errors follow this shape:
```json
{ "error": "description of the error" }
```

Common HTTP status codes:
| Code | Meaning |
|------|---------|
| 400 | Bad request / validation failure |
| 401 | Missing or invalid token |
| 403 | Insufficient role |
| 404 | Resource not found |
| 500 | Internal server error |

---

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `MONGODB_URI` | `mongodb://localhost:27017` | MongoDB connection string |
| `GIN_MODE` | `debug` | Set to `release` in production |

JWT secret is hardcoded as `stock_management_secret_key_2026` — replace before production.
