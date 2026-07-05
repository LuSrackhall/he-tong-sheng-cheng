# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: core-business.spec.ts >> Business flows >> 3.2 List assets
- Location: e2e/scenarios/core-business.spec.ts:91:3

# Error details

```
Error: expect(received).toBe(expected) // Object.is equality

Expected: true
Received: false
```

# Test source

```ts
  1   | import { test, expect } from '@playwright/test';
  2   | 
  3   | const API = 'http://localhost:8080';
  4   | let token = '';
  5   | let assetId = 0;
  6   | let tenantId = 0;
  7   | let contractId = 0;
  8   | 
  9   | async function login() {
  10  |   const resp = await fetch(`${API}/api/auth/login`, {
  11  |     method: 'POST',
  12  |     headers: { 'Content-Type': 'application/json' },
  13  |     body: JSON.stringify({ username: 'admin', password: 'admin123' }),
  14  |   });
  15  |   const data = await resp.json();
  16  |   token = data.token;
  17  |   return data;
  18  | }
  19  | 
  20  | // ============================================================
  21  | // 1. Health & Connectivity (no auth needed)
  22  | // ============================================================
  23  | test('1.1 Health check returns ok', async () => {
  24  |   const resp = await fetch(`${API}/api/health`);
  25  |   expect(resp.status).toBe(200);
  26  |   const data = await resp.json();
  27  |   expect(data.status).toBe('ok');
  28  | });
  29  | 
  30  | // ============================================================
  31  | // 2. Authentication (independent tests)
  32  | // ============================================================
  33  | test('2.1 Login with valid credentials returns token', async () => {
  34  |   const data = await login();
  35  |   expect(data.token).toBeDefined();
  36  |   expect(typeof data.token).toBe('string');
  37  | });
  38  | 
  39  | test('2.2 Login with wrong password returns 401', async () => {
  40  |   const resp = await fetch(`${API}/api/auth/login`, {
  41  |     method: 'POST',
  42  |     headers: { 'Content-Type': 'application/json' },
  43  |     body: JSON.stringify({ username: 'admin', password: 'wrong' }),
  44  |   });
  45  |   expect(resp.status).toBe(401);
  46  | });
  47  | 
  48  | test('2.3 Login with empty password returns 400', async () => {
  49  |   const resp = await fetch(`${API}/api/auth/login`, {
  50  |     method: 'POST',
  51  |     headers: { 'Content-Type': 'application/json' },
  52  |     body: JSON.stringify({ username: 'admin', password: '' }),
  53  |   });
  54  |   expect(resp.status).toBe(400);
  55  | });
  56  | 
  57  | test('2.4 No auth token returns 401', async () => {
  58  |   const resp = await fetch(`${API}/api/contracts`);
  59  |   expect(resp.status).toBe(401);
  60  | });
  61  | 
  62  | test('2.5 Auth/me returns user info', async () => {
  63  |   await login();
  64  |   const resp = await fetch(`${API}/api/auth/me`, {
  65  |     headers: { Authorization: `Bearer ${token}` },
  66  |   });
  67  |   expect(resp.status).toBe(200);
  68  |   const data = await resp.json();
  69  |   expect(data.username).toBe('admin');
  70  | });
  71  | 
  72  | // ============================================================
  73  | // 3-8. Shared tests (run serially to share state)
  74  | // ============================================================
  75  | test.describe.serial('Business flows', () => {
  76  | 
  77  |   // --- 3. Asset Management ---
  78  |   test('3.1 Create asset', async () => {
  79  |     await login();
  80  |     const resp = await fetch(`${API}/api/assets`, {
  81  |       method: 'POST',
  82  |       headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
  83  |       body: JSON.stringify({ name: 'E2E测试商铺', monthlyRent: 5000, type: '商铺', area: 80 }),
  84  |     });
  85  |     expect(resp.status).toBe(201);
  86  |     const data = await resp.json();
  87  |     expect(data.id).toBeDefined();
  88  |     assetId = data.id;
  89  |   });
  90  | 
  91  |   test('3.2 List assets', async () => {
  92  |     await login();
  93  |     const resp = await fetch(`${API}/api/assets`, {
  94  |       headers: { Authorization: `Bearer ${token}` },
  95  |     });
  96  |     expect(resp.status).toBe(200);
  97  |     const data = await resp.json();
> 98  |     expect(Array.isArray(data)).toBe(true);
      |                                 ^ Error: expect(received).toBe(expected) // Object.is equality
  99  |   });
  100 | 
  101 |   test('3.3 Search assets', async () => {
  102 |     await login();
  103 |     const resp = await fetch(`${API}/api/assets?search=E2E`, {
  104 |       headers: { Authorization: `Bearer ${token}` },
  105 |     });
  106 |     expect(resp.status).toBe(200);
  107 |   });
  108 | 
  109 |   test('3.4 Get asset by ID', async () => {
  110 |     await login();
  111 |     const resp = await fetch(`${API}/api/assets/${assetId}`, {
  112 |       headers: { Authorization: `Bearer ${token}` },
  113 |     });
  114 |     expect(resp.status).toBe(200);
  115 |   });
  116 | 
  117 |   test('3.5 Update asset', async () => {
  118 |     await login();
  119 |     const resp = await fetch(`${API}/api/assets/${assetId}`, {
  120 |       method: 'PATCH',
  121 |       headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
  122 |       body: JSON.stringify({ monthlyRent: 6000 }),
  123 |     });
  124 |     expect(resp.status).toBe(200);
  125 |   });
  126 | 
  127 |   test('3.6 Get nonexistent asset returns 404', async () => {
  128 |     await login();
  129 |     const resp = await fetch(`${API}/api/assets/99999`, {
  130 |       headers: { Authorization: `Bearer ${token}` },
  131 |     });
  132 |     expect(resp.status).toBe(404);
  133 |   });
  134 | 
  135 |   // --- 4. Tenant Management ---
  136 |   test('4.1 Create tenant', async () => {
  137 |     await login();
  138 |     const resp = await fetch(`${API}/api/tenants`, {
  139 |       method: 'POST',
  140 |       headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
  141 |       body: JSON.stringify({ name: 'E2E测试租户', phone: '13800138000', idCard: '110101199001011234' }),
  142 |     });
  143 |     expect(resp.status).toBe(201);
  144 |     const data = await resp.json();
  145 |     expect(data.id).toBeDefined();
  146 |     tenantId = data.id;
  147 |   });
  148 | 
  149 |   test('4.2 List tenants', async () => {
  150 |     await login();
  151 |     const resp = await fetch(`${API}/api/tenants`, {
  152 |       headers: { Authorization: `Bearer ${token}` },
  153 |     });
  154 |     expect(resp.status).toBe(200);
  155 |   });
  156 | 
  157 |   test('4.3 Get tenant by ID', async () => {
  158 |     await login();
  159 |     const resp = await fetch(`${API}/api/tenants/${tenantId}`, {
  160 |       headers: { Authorization: `Bearer ${token}` },
  161 |     });
  162 |     expect(resp.status).toBe(200);
  163 |   });
  164 | 
  165 |   test('4.4 Update tenant', async () => {
  166 |     await login();
  167 |     const resp = await fetch(`${API}/api/tenants/${tenantId}`, {
  168 |       method: 'PATCH',
  169 |       headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
  170 |       body: JSON.stringify({ phone: '13900139000' }),
  171 |     });
  172 |     expect(resp.status).toBe(200);
  173 |   });
  174 | 
  175 |   // --- 5. Contract Management ---
  176 |   test('5.1 Create contract', async () => {
  177 |     await login();
  178 |     const resp = await fetch(`${API}/api/contracts`, {
  179 |       method: 'POST',
  180 |       headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
  181 |       body: JSON.stringify({
  182 |         assetId, tenantId, startDate: '2026-01-01', endDate: '2027-12-31',
  183 |         monthlyRent: 5000, status: 'active',
  184 |       }),
  185 |     });
  186 |     const data = await resp.json();
  187 |     expect(data.id).toBeDefined();
  188 |     contractId = data.id;
  189 |   });
  190 | 
  191 |   test('5.2 Get contract by ID', async () => {
  192 |     await login();
  193 |     const resp = await fetch(`${API}/api/contracts/${contractId}`, {
  194 |       headers: { Authorization: `Bearer ${token}` },
  195 |     });
  196 |     expect(resp.status).toBe(200);
  197 |   });
  198 | 
```