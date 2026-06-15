# Frontend Specification — StockFlow

## Stack

| Layer | Technology |
|-------|-----------|
| Framework | React 18 |
| Build tool | Vite 5 |
| Routing | Manual state-based (`currentView` string) |
| Styling | CSS variables (dark/light theme, `index.css`) |
| i18n | Custom context (`i18n.jsx`) — English / Thai |
| State | Local `useState` per page, `localStorage` for session |

Dev server: `http://localhost:5174` (or 5173 if available)
API base: `http://localhost:8080`

---

## Project Structure

```
frontend/src/
├── main.jsx              # React entry point
├── App.jsx               # Root component — auth, routing, layout, theme
├── App.css               # Component-scoped styles
├── index.css             # Global CSS variables and base styles
├── i18n.jsx              # Translation context and strings (EN / TH)
└── pages/
    ├── Dashboard.jsx     # KPI overview and recent activity
    ├── Catalog.jsx       # Product catalog — CRUD
    ├── StockLevels.jsx   # Stock viewer and manual adjustment
    ├── Inbound.jsx       # Purchase orders and goods receiving
    ├── Outbound.jsx      # Sales orders, shipping, RMA returns
    ├── Analytics.jsx     # FIFO/LIFO/Average Cost valuation report
    ├── Reconciliation.jsx# Physical stock count and approval workflow
    └── BarcodeScanner.jsx# Barcode scan for stock in/out
```

---

## App.jsx — Root Component

Owns all global state and passes props down to pages.

### Session State

| State | Type | Storage | Description |
|-------|------|---------|-------------|
| `token` | string | `localStorage` | JWT token |
| `username` | string | `localStorage` | Logged-in username |
| `role` | string | `localStorage` | `Admin` / `WarehouseStaff` / `SalesRep` |

Session is restored from `localStorage` on page load.

### Global State

| State | Type | Description |
|-------|------|-------------|
| `currentView` | string | Active page key (see Navigation) |
| `theme` | string | `'dark'` or `'light'` |
| `toast` | `{message, visible}` | Toast notification — auto-hides after 4 seconds |
| `lang` | string | `'en'` or `'th'` (from i18n context) |

### Helpers passed as props

```js
// Auth headers for every fetch call
const apiHeaders = { Authorization: `Bearer ${token}` }

// Trigger a toast notification
const triggerToast = (message) => { /* shows toast for 4s */ }
```

### Login Flow

1. User submits username + password via login form
2. `POST /api/login` — on success, stores `token`, `username`, `role` in localStorage
3. On failure, shows inline error message
4. Logout clears localStorage and resets all session state

---

## Navigation

| `currentView` key | Page Component | Nav Label |
|-------------------|---------------|-----------|
| `dashboard` | `Dashboard` | Dashboard Overview |
| `catalog` | `Catalog` | Product Catalog |
| `stock` | `StockLevels` | Stock Levels |
| `inbound` | `Inbound` | Inbound (Procure) |
| `outbound` | `Outbound` | Outbound (Fulfill) |
| `analytics` | `Analytics` | Financial Valuation |
| `scanner` | `BarcodeScanner` | Barcode Scanner |
| `reconciliation` | `Reconciliation` | Reconciliation |

All pages receive: `{ userRole, apiHeaders, triggerToast }`

---

## Pages

### Dashboard (`Dashboard.jsx`)

**Purpose:** High-level KPI overview and quick links.

**Data fetched on load:**
- `GET /api/stock` — total products, low stock count
- `GET /api/inbound/po` — pending PO count
- `GET /api/outbound/so` — pending SO count
- `GET /api/audit` — recent activity feed

**KPI cards displayed:**
- Total Products
- Low Stock Alerts (availableStock ≤ reorderPoint)
- Pending Purchase Orders
- Pending Sales Orders

**Recent Activity:** Last 10 audit log entries (action, SKU, user, timestamp)

---

### Catalog (`Catalog.jsx`)

**Purpose:** View and manage the product catalog.

**Data fetched on load:**
- `GET /api/products`

**Features:**
- Search/filter products by name, SKU, or category
- Table view: SKU, Name, Category, Brand, Cost Price, Selling Price
- **Create product** modal — `POST /api/products`
  - Role required: `Admin` or `SalesRep`
- **Edit product** modal — `PUT /api/products/:id`
  - Role required: `Admin`
- **Delete product** — `DELETE /api/products/:id`
  - Role required: `Admin`

**Form fields (Create / Edit):**

| Field | Type | Required |
|-------|------|----------|
| SKU | text | Yes (Create only) |
| Name | text | Yes |
| Barcode | text | No |
| Category | text | No |
| Brand | text | No |
| Cost Price | number (> 0) | Yes |
| Selling Price | number (> 0) | Yes |
| Supplier IDs | comma-separated text | No |

---

### Stock Levels (`StockLevels.jsx`)

**Purpose:** View current stock and manually adjust quantities.

**Data fetched on load:**
- `GET /api/stock`
- `GET /api/products` (for product name lookup by ID)

**Features:**
- Search by SKU or product name
- Table view: SKU, Product Name, Stock On Hand, Committed, Available, Reorder Point, Location, Status badge
- Status badge: `LOW STOCK` (red) when `availableStock ≤ reorderPoint`, otherwise `OK` (green)
- **Adjust Stock** modal — `POST /api/stock/adjust`
  - Role required: `Admin` or `WarehouseStaff`

**Adjust Stock form fields:**

| Field | Type | Required | Notes |
|-------|------|----------|-------|
| Stock On Hand | number (≥ 0) | Yes | Sets absolute value |
| Reorder Point | number | Yes | |
| Warehouse | text | No | |
| Aisle | text | No | |
| Bin | text | No | |
| Note | text | Yes | Reason for adjustment |

---

### Inbound (`Inbound.jsx`)

**Purpose:** Manage purchase orders and record goods received.

**Data fetched on load:**
- `GET /api/inbound/po`
- `GET /api/products` (for item selection dropdown)

**Features:**

**PO List table:** PO Number, Supplier, Status badge, Total Amount, Created Date, Actions

Status badges: `Draft` (grey) → `Sent` (blue) → `Received` (green)

**Create PO** modal — `POST /api/inbound/po`
- Role required: `Admin` or `SalesRep`
- Dynamic line items: add/remove rows
- Each row: Product (dropdown), auto-fills SKU and Cost Price, editable Quantity

**Send PO** button — `POST /api/inbound/po/send/:id`
- Visible when status = `Draft`
- Role required: `Admin` or `SalesRep`

**Receive Goods** modal — `POST /api/inbound/po/receive/:id`
- Visible when status = `Sent`
- Role required: `Admin` or `WarehouseStaff`
- Shows each item with ordered qty and input for received qty this batch
- Supports partial receiving

---

### Outbound (`Outbound.jsx`)

**Purpose:** Manage sales orders, shipping, and returns (RMA).

**Tabs:** `Sales Orders` | `Returns (RMA)`

**Data fetched on load:**
- `GET /api/outbound/so`
- `GET /api/outbound/rma`
- `GET /api/products`
- `GET /api/stock` (for available stock display during SO creation)

#### Sales Orders Tab

**SO List table:** SO Number, Customer, Status badge, Total Amount, Created Date, Actions

Status badges: `Pending` (yellow) → `Shipped` (green)

**Create SO** modal — `POST /api/outbound/so`
- Role required: `Admin` or `SalesRep`
- Dynamic line items: add/remove rows
- Each row: Product (dropdown), auto-fills SKU and Selling Price, editable Quantity
- Shows available stock per product as a hint
- Server validates stock availability before creating

**Ship Order** button — `POST /api/outbound/so/ship/:id`
- Visible when status = `Pending`
- Role required: `Admin` or `WarehouseStaff`

#### Returns (RMA) Tab

**RMA List table:** Sales Order ID, SKU, Quantity, Reason, Status, Created Date

**Create RMA** modal — `POST /api/outbound/rma`
- Role required: `Admin` or `WarehouseStaff`

**RMA form fields:**

| Field | Type | Required | Notes |
|-------|------|----------|-------|
| Sales Order ID | text | Yes | ObjectId of the original SO |
| Product | dropdown | Yes | Selects productId and SKU |
| Quantity | number | Yes | > 0 |
| Reason | text | No | |
| Disposition | select | Yes | `Restocked` or `Damaged` |

---

### Analytics (`Analytics.jsx`)

**Purpose:** Financial inventory valuation report.

**Data fetched on load:**
- `GET /api/reports/valuation`

**Displays:**

Summary cards:
- Total FIFO Valuation
- Total LIFO Valuation
- Total Average Cost Valuation

Per-SKU table: SKU, Product Name, Stock On Hand, Base Cost, FIFO Value, LIFO Value, Avg Cost Value

Bar chart (SVG-rendered): Compares the three totals visually.

No write operations on this page.

---

### Reconciliation (`Reconciliation.jsx`)

**Purpose:** Physical stock count submission and approval.

**Views:** `History` | `New Count`

**Data fetched on load:**
- `GET /api/reconciliation`
- `GET /api/stock` (for expected quantities)

#### History View

Table: Date, Submitted By, Status badge, Note, Approved By, Actions

Status badges: `Draft` (yellow) | `Approved` (green)

**Approve** button — `POST /api/reconciliation/:id/approve`
- Visible when status = `Draft`
- Role required: `Admin`
- On approval, stock records are updated to match actual counts

#### New Count View

- Table listing every SKU with its current `stockOnHand` (expected) and an input for actual physical count
- Note field
- **Submit Count** — `POST /api/reconciliation`
- Role required: `Admin` or `WarehouseStaff`

---

### Barcode Scanner (`BarcodeScanner.jsx`)

**Purpose:** Fast stock in/out via barcode scan (USB scanner or manual entry).

**API used:**
- `POST /api/stock/scan`

**Features:**
- Auto-focuses input field so a USB barcode scanner can type directly
- Toggle: `Stock IN` (▲) / `Stock OUT` (▼)
- Fields: Barcode, Quantity, Note (optional)
- On success: shows product name, new stock level, quantity changed
- Recent Scans table: last 10 scans this session (barcode, product, status, qty, new stock, time)
- Role required: `Admin` or `WarehouseStaff` — others see access denied message

---

## i18n (`i18n.jsx`)

Simple custom context. Two languages: `en` (English) and `th` (Thai).

Usage in components:
```js
const { t } = useT();
// t('key') returns translated string
```

Language toggle is in the top bar. Preference is stored in component state (not persisted).

Translation keys cover: nav labels, login form, scanner UI, reconciliation labels, and general UI strings.

---

## Theme

Two themes: `dark` (default) and `light`. Toggled via a button in the top bar.

CSS custom properties are set on `:root` and swapped by a `.light` class on `body`:

Key variables:
```css
--bg-primary      /* main background */
--bg-secondary    /* card/panel background */
--bg-tertiary     /* input/table row background */
--text-primary    /* main text */
--text-muted      /* secondary text */
--accent          /* primary action color (blue) */
--danger          /* red — delete, low stock */
--success         /* green — shipped, received, ok */
--warning         /* yellow — pending, draft */
```

---

## Role-Based UI Rules

| Feature | Admin | WarehouseStaff | SalesRep |
|---------|-------|---------------|---------|
| View all pages | ✅ | ✅ | ✅ |
| Create / Edit product | ✅ | ❌ | ✅ (create only) |
| Delete product | ✅ | ❌ | ❌ |
| Adjust stock | ✅ | ✅ | ❌ |
| Barcode scan | ✅ | ✅ | ❌ |
| Create PO | ✅ | ❌ | ✅ |
| Send PO | ✅ | ❌ | ✅ |
| Receive PO | ✅ | ✅ | ❌ |
| Create SO | ✅ | ❌ | ✅ |
| Ship SO | ✅ | ✅ | ❌ |
| Create RMA | ✅ | ✅ | ❌ |
| Submit reconciliation | ✅ | ✅ | ❌ |
| Approve reconciliation | ✅ | ❌ | ❌ |

Role checks happen both server-side (Gin middleware) and client-side (button visibility / toast error).
