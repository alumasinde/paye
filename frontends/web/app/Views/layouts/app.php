<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
<meta name="theme-color" content="#101828">
<title><?= h($title ?? 'Budget254 Payroll') ?></title>
<link rel="stylesheet" href="/assets/app.css">
</head>
<body>
<div class="shell">
    <aside class="sidebar">
        <a class="brand" href="/dashboard"><span class="brand-mark">B</span><span>Budget254</span></a>
        <div class="workspace-label">PAYROLL WORKSPACE</div>
        <nav>
            <a class="nav-item <?= ($activeNav ?? '') === 'dashboard' ? 'active' : '' ?>" href="/dashboard">Dashboard</a>
            <a class="nav-item <?= ($activeNav ?? '') === 'companies' ? 'active' : '' ?>" href="/companies">Companies</a>
            <span class="nav-item disabled">Employees</span>
            <span class="nav-item disabled">Payroll</span>
            <span class="nav-item disabled">Reports</span>
        </nav>
        <div class="sidebar-note">Build your workspace one step at a time. Companies can have their own employees and payroll runs.</div>
    </aside>
    <main class="main">
        <header class="topbar app-topbar">
            <div>
                <p class="eyebrow">BUDGET254 PAYROLL</p>
                <h1><?= h($title ?? 'Payroll workspace') ?></h1>
            </div>
            <div class="account-menu">
                <span><?= h((string)($user['first_name'] ?? 'Account')) ?></span>
                <form method="post" action="/logout">
                    <input type="hidden" name="_csrf" value="<?= h(csrf_token()) ?>">
                    <button class="button secondary small-button" type="submit">Log out</button>
                </form>
            </div>
        </header>
        <?php require $viewFile; ?>
    </main>
</div>
</body>
</html>
