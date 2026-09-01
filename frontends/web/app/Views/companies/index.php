<?php
$success = Session::flash('success');
$error = Session::flash('error');
?>
<?php if ($success): ?><div class="alert success"><?= h($success) ?></div><?php endif; ?>
<?php if ($error): ?><div class="alert error"><?= h($error) ?></div><?php endif; ?>
<?php if ($loadError): ?><div class="alert error"><?= h($loadError) ?></div><?php endif; ?>

<section class="page-heading">
    <div>
        <p class="eyebrow">WORKSPACES</p>
        <h2>Your companies</h2>
        <p class="muted">Each company has its own employees, payroll runs and payroll records.</p>
    </div>
    <a class="button link-button" href="/companies/create">Create company</a>
</section>

<?php if ($companies): ?>
<div class="company-grid">
<?php foreach ($companies as $company): ?>
    <article class="company-card card">
        <div class="company-card-top">
            <div>
                <span class="company-initial"><?= h(strtoupper(substr((string)($company['legal_name'] ?? 'C'), 0, 1))) ?></span>
                <h3><?= h((string)($company['legal_name'] ?? 'Company')) ?></h3>
                <?php if (!empty($company['trading_name'])): ?><p class="muted"><?= h((string)$company['trading_name']) ?></p><?php endif; ?>
            </div>
            <span class="status status-approved"><?= h((string)($company['status'] ?? 'ACTIVE')) ?></span>
        </div>
        <dl class="company-meta">
            <div><dt>KRA PIN</dt><dd><?= h((string)($company['kra_pin'] ?? '—')) ?></dd></div>
            <div><dt>Payroll</dt><dd><?= h((string)($company['payroll_frequency'] ?? 'MONTHLY')) ?></dd></div>
            <div><dt>Currency</dt><dd><?= h((string)($company['currency_code'] ?? 'KES')) ?></dd></div>
        </dl>
        <div class="company-card-footer">
            <span class="muted">Employee and payroll tools are next.</span>
        </div>
    </article>
<?php endforeach; ?>
</div>
<?php else: ?>
<section class="empty card">
    <p class="eyebrow">GET STARTED</p>
    <h2>Create your first company</h2>
    <p class="muted">Add your legal company details first. This keeps employees and payroll data separated as your account grows.</p>
    <a class="button link-button" href="/companies/create">Create company</a>
</section>
<?php endif; ?>
