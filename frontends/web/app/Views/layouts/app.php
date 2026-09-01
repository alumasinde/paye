<?php $companyContext = company_context(); ?>
<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
<meta name="theme-color" content="<?= h($companyContext['primary_color']) ?>">
<title><?= h($title ?? 'Budget254 Payroll') ?></title>
<link rel="stylesheet" href="/assets/app.css">
</head>
<body style="--brand:<?= h($companyContext['primary_color']) ?>;--brand-strong:<?= h($companyContext['secondary_color']) ?>">
<div class="shell">
    <aside class="sidebar">
        <div class="sidebar-brand-row">
            <a class="brand" href="/dashboard"><span class="brand-mark">B</span><span>Budget254</span></a>
        </div>

        <?php if ($companyContext['name'] !== ''): ?>
            <div class="company-context"><?= h($companyContext['name']) ?></div>
        <?php endif; ?>

        <nav class="sidebar-nav" aria-label="Main navigation">
            <a class="nav-item <?= ($activeNav ?? '') === 'dashboard' ? 'active' : '' ?>" href="/dashboard">Dashboard</a>
            <a class="nav-item <?= ($activeNav ?? '') === 'companies' ? 'active' : '' ?>" href="/companies">Companies</a>
            <a class="nav-item <?= ($activeNav ?? '') === 'employees' ? 'active' : '' ?>" href="/employees">Employees</a>
            <a class="nav-item <?= ($activeNav ?? '') === 'departments' ? 'active' : '' ?>" href="/departments">Departments</a>
            <a class="nav-item <?= ($activeNav ?? '') === 'payroll' ? 'active' : '' ?>" href="/payroll">Payroll</a>
            <a class="nav-item <?= ($activeNav ?? '') === 'reports' ? 'active' : '' ?>" href="/reports/payroll<?= $companyContext['id'] !== '' ? '?company=' . rawurlencode($companyContext['id']) : '' ?>">Reports</a>
        </nav>

        <div class="sidebar-bottom">
            <a class="nav-item <?= ($activeNav ?? '') === 'settings' ? 'active' : '' ?>" href="/settings<?= $companyContext['id'] !== '' ? '?company=' . rawurlencode($companyContext['id']) : '' ?>">Settings</a>
        </div>
    </aside>

    <main class="main">
        <header class="topbar app-topbar">
            <div class="topbar-title">
                <h1><?= h($title ?? 'Payroll workspace') ?></h1>
                <?php if ($companyContext['name'] !== ''): ?><span><?= h($companyContext['name']) ?></span><?php endif; ?>
            </div>
            <div class="account-menu">
                <span class="account-name"><?= h((string)($user['first_name'] ?? 'Account')) ?></span>
                <form method="post" action="/logout">
                    <input type="hidden" name="_csrf" value="<?= h(csrf_token()) ?>">
                    <button class="icon-button" type="submit" title="Log out">Log out</button>
                </form>
            </div>
        </header>
        <div class="content"><?php require $viewFile; ?></div>
    </main>
</div>
</body>
</html>
