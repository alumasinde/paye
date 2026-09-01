<?php
declare(strict_types=1);

$apiBase = rtrim(getenv('PAYROLL_API_BASE_URL') ?: 'http://127.0.0.1:8080/api/v1', '/');

function h(string $value): string { return htmlspecialchars($value, ENT_QUOTES, 'UTF-8'); }

function request(string $method, string $path, ?array $payload = null): array {
    global $apiBase;

    $headers = ['Accept: application/json'];
    $token = $_COOKIE['payroll_access_token'] ?? '';
    if ($token !== '') $headers[] = 'Authorization: Bearer ' . $token;
    if ($payload !== null) $headers[] = 'Content-Type: application/json';

    $ch = curl_init($apiBase . $path);
    curl_setopt_array($ch, [
        CURLOPT_CUSTOMREQUEST => $method,
        CURLOPT_HTTPHEADER => $headers,
        CURLOPT_RETURNTRANSFER => true,
        CURLOPT_TIMEOUT => 20,
    ]);
    if ($payload !== null) curl_setopt($ch, CURLOPT_POSTFIELDS, json_encode($payload));

    $raw = curl_exec($ch);
    $status = (int) curl_getinfo($ch, CURLINFO_RESPONSE_CODE);
    $error = curl_error($ch);
    curl_close($ch);

    if ($raw === false || $error !== '') return ['ok' => false, 'status' => 0, 'message' => 'Could not reach the payroll service.'];
    $body = json_decode($raw, true);
    if (!is_array($body)) $body = [];
    return [
        'ok' => $status >= 200 && $status < 300,
        'status' => $status,
        'data' => $body,
        'message' => (string)($body['message'] ?? $body['error']['message'] ?? 'Request could not be completed.'),
    ];
}

$companyID = trim((string)($_GET['company'] ?? ''));
$runID = trim((string)($_GET['run'] ?? ''));
$message = '';
$error = '';

if ($_SERVER['REQUEST_METHOD'] === 'POST') {
    $companyID = trim((string)($_POST['company'] ?? $companyID));
    $runID = trim((string)($_POST['run'] ?? $runID));
    $action = strtoupper(trim((string)($_POST['action'] ?? '')));
    $paths = [
        'CALCULATE' => 'calculate',
        'REVIEW' => 'review',
        'APPROVE' => 'approve',
        'FINALIZE' => 'finalize',
        'LOCK' => 'lock',
    ];

    if ($companyID === '' || $runID === '' || !isset($paths[$action])) {
        $error = 'Select a company and payroll run before continuing.';
    } else {
        $response = request('POST', '/companies/' . rawurlencode($companyID) . '/payroll-runs/' . rawurlencode($runID) . '/' . $paths[$action]);
        if ($response['ok']) {
            $message = $action === 'CALCULATE' ? 'Payroll calculation completed.' : 'Payroll moved to the next workflow stage.';
        } else {
            $error = $response['message'];
        }
    }
}

$companiesResponse = request('GET', '/companies');
$companies = $companiesResponse['ok'] ? (array)($companiesResponse['data']['companies'] ?? []) : [];

if ($companyID === '' && count($companies) === 1) $companyID = (string)($companies[0]['id'] ?? '');

$runs = [];
if ($companyID !== '') {
    $runsResponse = request('GET', '/companies/' . rawurlencode($companyID) . '/payroll-runs');
    if ($runsResponse['ok']) $runs = (array)($runsResponse['data']['payroll_runs'] ?? []);
    elseif ($error === '') $error = $runsResponse['message'];
}

if ($runID === '' && count($runs) === 1) $runID = (string)($runs[0]['id'] ?? '');

$detail = null;
if ($companyID !== '' && $runID !== '') {
    $detailResponse = request('GET', '/companies/' . rawurlencode($companyID) . '/payroll-runs/' . rawurlencode($runID));
    if ($detailResponse['ok']) $detail = $detailResponse['data'];
    elseif ($error === '') $error = $detailResponse['message'];
}

$employees = $detail ? (array)($detail['employees'] ?? []) : [];
$totals = ['gross' => 0.0, 'paye' => 0.0, 'deductions' => 0.0, 'net' => 0.0];
$counts = ['CALCULATED' => 0, 'FAILED' => 0, 'PENDING' => 0];
foreach ($employees as $employee) {
    $totals['gross'] += (float)($employee['gross_salary'] ?? 0);
    $totals['paye'] += (float)($employee['paye'] ?? 0);
    $totals['deductions'] += (float)($employee['total_deductions'] ?? 0);
    $totals['net'] += (float)($employee['net_salary'] ?? 0);
    $status = (string)($employee['status'] ?? 'PENDING');
    $counts[$status] = ($counts[$status] ?? 0) + 1;
}

function money(float $value): string { return 'KES ' . number_format($value, 2); }
function employeeName(array $employee): string {
    return trim(implode(' ', array_filter([(string)($employee['first_name'] ?? ''), (string)($employee['middle_name'] ?? ''), (string)($employee['last_name'] ?? '')])));
}

$status = (string)($detail['status'] ?? '');
$actions = [
    'DRAFT' => ['CALCULATE' => ['Calculate payroll', 'Calculate all employee payroll figures']],
    'CALCULATED' => ['REVIEW' => ['Send for review', 'Move this payroll into the review stage']],
    'REVIEW' => ['APPROVE' => ['Approve payroll', 'Confirm this payroll is ready for finalization']],
    'APPROVED' => ['FINALIZE' => ['Finalize payroll', 'Mark this payroll as an official completed payroll']],
    'FINALIZED' => ['LOCK' => ['Lock payroll', 'Prevent further workflow changes']],
];
$availableActions = $actions[$status] ?? [];
?><!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
<title>Payroll review | Budget254</title>
<link rel="stylesheet" href="assets/app.css">
</head>
<body>
<div class="shell">
  <aside class="sidebar">
    <a class="brand" href="index.php"><span class="brand-mark">B</span><span>Budget254</span></a>
    <div class="workspace-label">PAYROLL WORKSPACE</div>
    <nav>
      <a class="nav-item active" href="index.php">Overview</a>
      <a class="nav-item" href="#employees">Employees</a>
      <a class="nav-item" href="#workflow">Payroll workflow</a>
    </nav>
    <div class="sidebar-note">Clear totals. Controlled workflow. One source of truth.</div>
  </aside>

  <main class="main">
    <header class="topbar">
      <div>
        <p class="eyebrow">PAYROLL</p>
        <h1>Review and finalize payroll</h1>
        <p class="muted">Check employee results, identify exceptions and move payroll forward with confidence.</p>
      </div>
      <?php if ($detail): ?><div class="status status-<?= strtolower(h($status)) ?>"><?= h($status) ?></div><?php endif; ?>
    </header>

    <?php if ($message !== ''): ?><div class="alert success"><?= h($message) ?></div><?php endif; ?>
    <?php if ($error !== ''): ?><div class="alert error"><?= h($error) ?></div><?php endif; ?>

    <section class="card selector-card">
      <form method="get" class="selectors">
        <label>Company
          <select name="company" onchange="this.form.submit()">
            <option value="">Select company</option>
            <?php foreach ($companies as $company): ?>
              <option value="<?= h((string)$company['id']) ?>" <?= $companyID === (string)$company['id'] ? 'selected' : '' ?>><?= h((string)($company['trading_name'] ?: $company['legal_name'])) ?></option>
            <?php endforeach; ?>
          </select>
        </label>
        <label>Payroll run
          <select name="run" onchange="this.form.submit()" <?= $companyID === '' ? 'disabled' : '' ?>>
            <option value="">Select payroll run</option>
            <?php foreach ($runs as $run): ?>
              <option value="<?= h((string)$run['id']) ?>" <?= $runID === (string)$run['id'] ? 'selected' : '' ?>><?= h((string)$run['period']) ?> · <?= h((string)$run['status']) ?></option>
            <?php endforeach; ?>
          </select>
        </label>
      </form>
    </section>

    <?php if (!$detail): ?>
      <section class="empty card"><h2>Select a payroll run</h2><p>Choose a company and payroll run to see totals, employee results and the next workflow action.</p></section>
    <?php else: ?>
      <section class="hero card">
        <div>
          <p class="eyebrow">PAYROLL PERIOD</p>
          <h2><?= h((string)$detail['period']) ?></h2>
          <p class="muted"><?= h(substr((string)$detail['period_start'], 0, 10)) ?> to <?= h(substr((string)$detail['period_end'], 0, 10)) ?></p>
        </div>
        <div class="hero-number"><span>Employees</span><strong><?= count($employees) ?></strong></div>
      </section>

      <section class="metrics">
        <article class="metric card"><span>Gross payroll</span><strong><?= money($totals['gross']) ?></strong></article>
        <article class="metric card"><span>PAYE</span><strong><?= money($totals['paye']) ?></strong></article>
        <article class="metric card"><span>Total deductions</span><strong><?= money($totals['deductions']) ?></strong></article>
        <article class="metric card accent"><span>Net payroll</span><strong><?= money($totals['net']) ?></strong></article>
      </section>

      <section id="workflow" class="workflow card">
        <div class="section-heading"><div><p class="eyebrow">WORKFLOW</p><h2>What happens next</h2></div></div>
        <div class="steps">
          <?php foreach (['DRAFT','CALCULATED','REVIEW','APPROVED','FINALIZED','LOCKED'] as $step): ?>
            <div class="step <?= $step === $status ? 'current' : '' ?>"><?= h($step) ?></div>
          <?php endforeach; ?>
        </div>
        <div class="workflow-action">
          <div>
            <?php if ($availableActions): foreach ($availableActions as $action => $copy): ?>
              <h3><?= h($copy[0]) ?></h3><p class="muted"><?= h($copy[1]) ?></p>
            <?php endforeach; else: ?><h3><?= $status === 'LOCKED' ? 'Payroll locked' : 'No action available' ?></h3><p class="muted"><?= $status === 'LOCKED' ? 'This payroll has completed its workflow.' : 'Resolve any calculation failures before moving forward.' ?></p><?php endif; ?>
          </div>
          <?php foreach ($availableActions as $action => $copy): ?>
            <form method="post">
              <input type="hidden" name="company" value="<?= h($companyID) ?>"><input type="hidden" name="run" value="<?= h($runID) ?>"><input type="hidden" name="action" value="<?= h($action) ?>">
              <button class="button" type="submit"><?= h($copy[0]) ?></button>
            </form>
          <?php endforeach; ?>
        </div>
      </section>

      <section id="employees" class="card table-card">
        <div class="section-heading"><div><p class="eyebrow">EMPLOYEE RESULTS</p><h2>Payroll review</h2><p class="muted"><?= ($counts['CALCULATED'] ?? 0) ?> calculated · <?= ($counts['FAILED'] ?? 0) ?> failed · <?= ($counts['PENDING'] ?? 0) ?> pending</p></div><input id="employee-search" type="search" placeholder="Search employee"></div>
        <div class="table-wrap">
          <table>
            <thead><tr><th>Employee</th><th>Number</th><th>Status</th><th>Gross</th><th>PAYE</th><th>Deductions</th><th>Net pay</th></tr></thead>
            <tbody id="employee-table">
              <?php foreach ($employees as $employee): ?><tr data-search="<?= h(strtolower(employeeName($employee) . ' ' . ($employee['employee_number'] ?? ''))) ?>">
                <td><strong><?= h(employeeName($employee)) ?></strong><?php if (!empty($employee['error_message'])): ?><small class="row-error"><?= h((string)$employee['error_message']) ?></small><?php endif; ?></td>
                <td><?= h((string)($employee['employee_number'] ?? '—')) ?></td>
                <td><span class="status small status-<?= strtolower((string)($employee['status'] ?? 'pending')) ?>"><?= h((string)($employee['status'] ?? 'PENDING')) ?></span></td>
                <td><?= money((float)($employee['gross_salary'] ?? 0)) ?></td><td><?= money((float)($employee['paye'] ?? 0)) ?></td><td><?= money((float)($employee['total_deductions'] ?? 0)) ?></td><td class="net"><?= money((float)($employee['net_salary'] ?? 0)) ?></td>
              </tr><?php endforeach; ?>
            </tbody>
          </table>
        </div>
      </section>
    <?php endif; ?>
  </main>
</div>
<script>
const search=document.getElementById('employee-search');
if(search){search.addEventListener('input',()=>{const q=search.value.toLowerCase().trim();document.querySelectorAll('#employee-table tr').forEach(row=>row.hidden=q!==''&&!row.dataset.search.includes(q));});}
</script>
</body>
</html>