<?php
$error = Session::flash('error');
$old = $old ?? [];
?>
<div class="page-heading">
    <div>
        <a class="back-link" href="/companies">← Back to companies</a>
        <p class="eyebrow">COMPANY SETUP</p>
        <h2>Create a company</h2>
        <p class="muted">Use the legal details used for payroll and statutory records. You can update most company settings later.</p>
    </div>
</div>

<?php if ($error): ?><div class="alert error"><?= h($error) ?></div><?php endif; ?>

<form method="post" action="/companies" class="setup-card card" novalidate>
<input type="hidden" name="_csrf" value="<?= h(csrf_token()) ?>">
<section class="form-section">
    <div><p class="eyebrow">01 · COMPANY IDENTITY</p><h3>Business details</h3></div>
    <div class="form-grid">
        <label>Legal company name<input type="text" name="legal_name" value="<?= h((string)($old['legal_name'] ?? '')) ?>" required></label>
        <label>Trading name <span class="optional">Optional</span><input type="text" name="trading_name" value="<?= h((string)($old['trading_name'] ?? '')) ?>"></label>
        <label>KRA PIN<input type="text" name="kra_pin" value="<?= h((string)($old['kra_pin'] ?? '')) ?>" autocapitalize="characters" required></label>
        <label>Company email<input type="email" name="email" value="<?= h((string)($old['email'] ?? '')) ?>" required></label>
        <label>Phone <span class="optional">Optional</span><input type="tel" name="phone" value="<?= h((string)($old['phone'] ?? '')) ?>"></label>
    </div>
</section>

<section class="form-section">
    <div><p class="eyebrow">02 · PAYROLL DEFAULTS</p><h3>How this company runs payroll</h3></div>
    <div class="form-grid">
        <label>Country code<input type="text" name="country_code" maxlength="2" value="<?= h((string)($old['country_code'] ?? 'KE')) ?>" required></label>
        <label>Currency code<input type="text" name="currency_code" maxlength="3" value="<?= h((string)($old['currency_code'] ?? 'KES')) ?>" required></label>
        <label>Payroll frequency
            <select name="payroll_frequency" required>
                <?php $frequency = (string)($old['payroll_frequency'] ?? 'MONTHLY'); ?>
                <option value="MONTHLY" <?= $frequency === 'MONTHLY' ? 'selected' : '' ?>>Monthly</option>
                <option value="WEEKLY" <?= $frequency === 'WEEKLY' ? 'selected' : '' ?>>Weekly</option>
                <option value="BIWEEKLY" <?= $frequency === 'BIWEEKLY' ? 'selected' : '' ?>>Bi-weekly</option>
            </select>
        </label>
    </div>
</section>

<div class="form-actions">
    <a class="button secondary link-button" href="/companies">Cancel</a>
    <button class="button" type="submit">Create company</button>
</div>
</form>
