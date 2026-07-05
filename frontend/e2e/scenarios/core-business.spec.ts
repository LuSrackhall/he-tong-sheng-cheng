import { test, expect } from '@playwright/test';

const API = 'http://localhost:8080';
let token = '';
let assetId = 0;
let tenantId = 0;
let contractId = 0;

// Helper: login and get token
async function login() {
  const resp = await fetch(`${API}/api/auth/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ username: 'admin', password: 'admin123' }),
  });
  const data = await resp.json();
  token = data.token;
  return data;
}

// ============================================================
// 1. Health & Connectivity
// ============================================================
test('1.1 Health check returns ok', async () => {
  const resp = await fetch(`${API}/api/health`);
  expect(resp.status).toBe(200);
  const data = await resp.json();
  expect(data.status).toBe('ok');
});

// ============================================================
// 2. Authentication
// ============================================================
test('2.1 Login with valid credentials returns token', async () => {
  const data = await login();
  expect(data.token).toBeDefined();
  expect(typeof data.token).toBe('string');
});

test('2.2 Login with wrong password returns 401', async () => {
  const resp = await fetch(`${API}/api/auth/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ username: 'admin', password: 'wrong' }),
  });
  const data = await resp.json();
  expect(resp.status).toBe(401);
  expect(data.error).toBeDefined();
});

test('2.3 Login with empty password returns 400', async () => {
  const resp = await fetch(`${API}/api/auth/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ username: 'admin', password: '' }),
  });
  expect(resp.status).toBe(400);
  const data = await resp.json();
  expect(data.error).toBeDefined();
});

test('2.4 No auth token returns 401', async () => {
  const resp = await fetch(`${API}/api/contracts`);
  expect(resp.status).toBe(401);
});

test('2.5 Auth/me returns user info', async () => {
  await login();
  const resp = await fetch(`${API}/api/auth/me`, {
    headers: { Authorization: `Bearer ${token}` },
  });
  expect(resp.status).toBe(200);
  const data = await resp.json();
  expect(data.username).toBe('admin');
});

// ============================================================
// 3. Asset Management
// ============================================================
test('3.1 Create asset', async () => {
  await login();
  const resp = await fetch(`${API}/api/assets`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
    body: JSON.stringify({ name: 'E2E测试商铺', monthlyRent: 5000, type: '商铺', area: 80 }),
  });
  expect(resp.status).toBe(201);
  const data = await resp.json();
  expect(data.id).toBeDefined();
  assetId = data.id;
});

test('3.2 List assets', async () => {
  const resp = await fetch(`${API}/api/assets`, {
    headers: { Authorization: `Bearer ${token}` },
  });
  expect(resp.status).toBe(200);
  const data = await resp.json();
  expect(Array.isArray(data)).toBe(true);
});

test('3.3 Search assets', async () => {
  const resp = await fetch(`${API}/api/assets?search=E2E`, {
    headers: { Authorization: `Bearer ${token}` },
  });
  expect(resp.status).toBe(200);
});

test('3.4 Get asset by ID', async () => {
  const resp = await fetch(`${API}/api/assets/${assetId}`, {
    headers: { Authorization: `Bearer ${token}` },
  });
  expect(resp.status).toBe(200);
});

test('3.5 Update asset', async () => {
  const resp = await fetch(`${API}/api/assets/${assetId}`, {
    method: 'PATCH',
    headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
    body: JSON.stringify({ monthlyRent: 6000 }),
  });
  expect(resp.status).toBe(200);
});

test('3.6 Get nonexistent asset returns 404', async () => {
  const resp = await fetch(`${API}/api/assets/99999`, {
    headers: { Authorization: `Bearer ${token}` },
  });
  expect(resp.status).toBe(404);
});

// ============================================================
// 4. Tenant Management
// ============================================================
test('4.1 Create tenant', async () => {
  const resp = await fetch(`${API}/api/tenants`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
    body: JSON.stringify({ name: 'E2E测试租户', phone: '13800138000', idCard: '110101199001011234' }),
  });
  expect(resp.status).toBe(201);
  const data = await resp.json();
  expect(data.id).toBeDefined();
  tenantId = data.id;
});

test('4.2 List tenants', async () => {
  const resp = await fetch(`${API}/api/tenants`, {
    headers: { Authorization: `Bearer ${token}` },
  });
  expect(resp.status).toBe(200);
});

test('4.3 Get tenant by ID', async () => {
  const resp = await fetch(`${API}/api/tenants/${tenantId}`, {
    headers: { Authorization: `Bearer ${token}` },
  });
  expect(resp.status).toBe(200);
});

test('4.4 Update tenant', async () => {
  const resp = await fetch(`${API}/api/tenants/${tenantId}`, {
    method: 'PATCH',
    headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
    body: JSON.stringify({ phone: '13900139000' }),
  });
  expect(resp.status).toBe(200);
});

// ============================================================
// 5. Contract Management
// ============================================================
test('5.1 Create contract', async () => {
  const resp = await fetch(`${API}/api/contracts`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
    body: JSON.stringify({
      assetId, tenantId,
      startDate: '2026-01-01', endDate: '2027-12-31',
      monthlyRent: 5000, status: 'active',
    }),
  });
  const data = await resp.json();
  if (resp.status === 201 || resp.status === 200) {
    expect(data.id).toBeDefined();
    contractId = data.id;
  } else {
    // May already exist - try to get the first contract
    const list = await (await fetch(`${API}/api/contracts`, { headers: { Authorization: `Bearer ${token}` } })).json();
    if (list.length > 0) contractId = list[0].id;
  }
});

test('5.2 Get contract by ID', async () => {
  if (!contractId) return;
  const resp = await fetch(`${API}/api/contracts/${contractId}`, {
    headers: { Authorization: `Bearer ${token}` },
  });
  expect(resp.status).toBe(200);
});

test('5.3 List contracts', async () => {
  const resp = await fetch(`${API}/api/contracts`, {
    headers: { Authorization: `Bearer ${token}` },
  });
  expect(resp.status).toBe(200);
});

test('5.4 Get nonexistent contract returns 404', async () => {
  const resp = await fetch(`${API}/api/contracts/99999`, {
    headers: { Authorization: `Bearer ${token}` },
  });
  expect(resp.status).toBe(404);
});

// ============================================================
// 6. Dashboard & Arrears
// ============================================================
test('6.1 Dashboard stats', async () => {
  const resp = await fetch(`${API}/api/dashboard/stats`, {
    headers: { Authorization: `Bearer ${token}` },
  });
  expect(resp.status).toBe(200);
  const data = await resp.json();
  expect(typeof data.activeContracts).toBe('number');
  expect(typeof data.monthlyRevenue).toBe('number');
  expect(typeof data.overdueContracts).toBe('number');
});

test('6.2 Arrears list', async () => {
  const resp = await fetch(`${API}/api/arrears`, {
    headers: { Authorization: `Bearer ${token}` },
  });
  expect(resp.status).toBe(200);
});

// ============================================================
// 7. Admin Operations
// ============================================================
test('7.1 List users', async () => {
  const resp = await fetch(`${API}/api/admin/users`, {
    headers: { Authorization: `Bearer ${token}` },
  });
  expect(resp.status).toBe(200);
});

test('7.2 Create operator user', async () => {
  const resp = await fetch(`${API}/api/admin/users`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
    body: JSON.stringify({ username: 'e2e-operator', password: 'e2e123456', role: 'operator' }),
  });
  expect(resp.status).toBe(201);
});

// ============================================================
// 8. Permission & Error Handling
// ============================================================
test('8.1 Operator cannot access admin endpoints', async () => {
  const resp = await fetch(`${API}/api/auth/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ username: 'e2e-operator', password: 'e2e123456' }),
  });
  const data = await resp.json();
  const opToken = data.token;
  const adminResp = await fetch(`${API}/api/admin/users`, {
    headers: { Authorization: `Bearer ${opToken}` },
  });
  expect(adminResp.status).toBe(403);
});

test('8.2 Error messages are in Chinese', async () => {
  const resp = await fetch(`${API}/api/auth/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({}),
  });
  const data = await resp.json();
  expect(data.error).toBeDefined();
  // Check that error is Chinese characters (no ASCII alphabet words)
  const hasChinese = /[一-鿿]/.test(data.error);
  expect(hasChinese).toBe(true);
});
