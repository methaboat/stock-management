import { createContext, useContext, useState } from 'react';

const en = {
  // Nav
  nav_dashboard: 'Dashboard Overview',
  nav_catalog: 'Product Catalog',
  nav_stock: 'Stock Levels',
  nav_inbound: 'Inbound (Procure)',
  nav_outbound: 'Outbound (Fulfill)',
  nav_analytics: 'Financial Valuation',
  nav_scanner: 'Barcode Scanner',
  nav_reconciliation: 'Reconciliation',

  // Login
  login_title: 'StockFlow',
  login_subtitle: 'Enter credentials to access warehouse management portal.',
  login_username: 'Username',
  login_password: 'Password',
  login_submit: 'Sign In',
  login_signing_in: 'Authenticating...',
  login_error_required: 'Username and Password are required.',
  login_demo: 'Demo Access Credentials',
  login_welcome: 'Welcome back',

  // Topbar
  topbar_logout: 'Logout',

  // Scanner
  scanner_title: 'Scan Barcode / Product Code',
  scanner_in: '▲ Stock IN',
  scanner_out: '▼ Stock OUT',
  scanner_barcode_label: 'Barcode / Product Code',
  scanner_barcode_placeholder: 'Scan or type barcode...',
  scanner_qty: 'Quantity',
  scanner_note: 'Note (optional)',
  scanner_note_placeholder: 'e.g. Received from supplier',
  scanner_confirm: 'Confirm',
  scanner_processing: 'Processing...',
  scanner_recent: "Today's Scans",
  scanner_no_scans: 'No scans yet this session.',
  scanner_time: 'Time',
  scanner_product: 'Product',
  scanner_status: 'Status',
  scanner_qty_col: 'Qty',
  scanner_stock: 'Stock',
  scanner_result_product: 'Product',
  scanner_result_sku: 'SKU',
  scanner_result_stock: 'Stock on hand',

  // Reconciliation
  recon_history: 'History',
  recon_new: '+ New Count',
  recon_no_history: 'No reconciliations yet.',
  recon_by: 'by',
  recon_approve: 'Approve',
  recon_title: 'End-of-Day Physical Count',
  recon_sku: 'SKU',
  recon_product: 'Product',
  recon_expected: 'Expected (System)',
  recon_actual: 'Actual Count',
  recon_diff: 'Diff',
  recon_note: 'Note (optional)',
  recon_note_placeholder: 'e.g. End of day count by staff',
  recon_submit: 'Submit Count',
  recon_submitting: 'Submitting...',
  recon_cancel: 'Cancel',
  recon_approved: 'Approved end-of-day stock',
  recon_submitted: 'Reconciliation submitted successfully',
};

const th = {
  // Nav
  nav_dashboard: 'ภาพรวมแดชบอร์ด',
  nav_catalog: 'รายการสินค้า',
  nav_stock: 'ระดับสต็อก',
  nav_inbound: 'รับสินค้าเข้า',
  nav_outbound: 'จ่ายสินค้าออก',
  nav_analytics: 'รายงานมูลค่า',
  nav_scanner: 'สแกนบาร์โค้ด',
  nav_reconciliation: 'ตรวจนับสต็อก',

  // Login
  login_title: 'StockFlow',
  login_subtitle: 'กรุณาเข้าสู่ระบบเพื่อใช้งานระบบจัดการคลังสินค้า',
  login_username: 'ชื่อผู้ใช้',
  login_password: 'รหัสผ่าน',
  login_submit: 'เข้าสู่ระบบ',
  login_signing_in: 'กำลังตรวจสอบ...',
  login_error_required: 'กรุณากรอกชื่อผู้ใช้และรหัสผ่าน',
  login_demo: 'ข้อมูลเข้าสู่ระบบทดสอบ',
  login_welcome: 'ยินดีต้อนรับ',

  // Topbar
  topbar_logout: 'ออกจากระบบ',

  // Scanner
  scanner_title: 'สแกนบาร์โค้ด / รหัสสินค้า',
  scanner_in: '▲ รับสินค้าเข้า',
  scanner_out: '▼ จ่ายสินค้าออก',
  scanner_barcode_label: 'บาร์โค้ด / รหัสสินค้า',
  scanner_barcode_placeholder: 'สแกนหรือพิมพ์บาร์โค้ด...',
  scanner_qty: 'จำนวน',
  scanner_note: 'หมายเหตุ (ถ้ามี)',
  scanner_note_placeholder: 'เช่น รับจากซัพพลายเออร์',
  scanner_confirm: 'ยืนยัน',
  scanner_processing: 'กำลังประมวลผล...',
  scanner_recent: 'สแกนวันนี้',
  scanner_no_scans: 'ยังไม่มีการสแกนในเซสชันนี้',
  scanner_time: 'เวลา',
  scanner_product: 'สินค้า',
  scanner_status: 'สถานะ',
  scanner_qty_col: 'จำนวน',
  scanner_stock: 'สต็อก',
  scanner_result_product: 'สินค้า',
  scanner_result_sku: 'รหัส SKU',
  scanner_result_stock: 'สต็อกคงเหลือ',

  // Reconciliation
  recon_history: 'ประวัติ',
  recon_new: '+ ตรวจนับใหม่',
  recon_no_history: 'ยังไม่มีการตรวจนับ',
  recon_by: 'โดย',
  recon_approve: 'อนุมัติ',
  recon_title: 'ตรวจนับสต็อกประจำวัน',
  recon_sku: 'รหัส SKU',
  recon_product: 'สินค้า',
  recon_expected: 'จำนวนในระบบ',
  recon_actual: 'จำนวนที่นับได้',
  recon_diff: 'ส่วนต่าง',
  recon_note: 'หมายเหตุ (ถ้ามี)',
  recon_note_placeholder: 'เช่น ตรวจนับสิ้นวันโดยพนักงาน',
  recon_submit: 'บันทึกการนับ',
  recon_submitting: 'กำลังบันทึก...',
  recon_cancel: 'ยกเลิก',
  recon_approved: 'อนุมัติสต็อกสิ้นวันแล้ว',
  recon_submitted: 'บันทึกการตรวจนับสำเร็จ',
};

const translations = { en, th };

export const LangContext = createContext({ lang: 'th', setLang: () => {}, t: k => k });

export function LangProvider({ children }) {
  const [lang, setLang] = useState('th');
  const t = key => translations[lang][key] ?? translations['en'][key] ?? key;
  return (
    <LangContext.Provider value={{ lang, setLang, t }}>
      {children}
    </LangContext.Provider>
  );
}

export const useT = () => useContext(LangContext);
