<?php
$success = Session::flash('success');
$error = Session::flash('error');
$primary = (string)($company['primary_color'] ?? '#15803D');
$secondary = (string)($company['secondary_color'] ?? '#166534');
?>
<?php if ($success): ?><div class="alert success"><?= h($success) ?></div><?php endif; ?>
<?php if ($error): ?><div class="alert error"><?= h($error) ?></div><?php endif; ?>
<?php if ($loadError): ?><div class="alert error"><?= h($loadError) ?></div><?php endif; ?>

<section class="page-heading compact-heading">
    <div>
        <p class="eyebrow">WORKSPACE SETTINGS</p>
        <h2>Appearance</h2>
    </div>
</section>

<?php if ($companies): ?>
<section class="settings-grid">
    <article class="card settings-card">
        <form method="get" class="company-switcher">
            <label>Company
                <select name="company" onchange="this.form.submit()">
                    <?php foreach ($companies as $item): $id=(string)($item['id'] ?? ''); ?>
                        <option value="<?= h($id) ?>" <?= $id===$companyId?'selected':'' ?>><?= h((string)($item['legal_name'] ?? 'Company')) ?></option>
                    <?php endforeach; ?>
                </select>
            </label>
        </form>

        <?php if ($company): ?>
        <form method="post" action="/settings/theme" class="settings-form">
            <input type="hidden" name="_csrf" value="<?= h(csrf_token()) ?>">
            <input type="hidden" name="company_id" value="<?= h($companyId) ?>">

            <div class="settings-section">
                <div>
                    <h3>Company theme</h3>
                    <p class="muted">Choose the colors used across this company workspace.</p>
                </div>

                <div class="theme-options">
                    <label class="color-field">
                        <span>Primary color</span>
                        <input type="color" name="primary_color" value="<?= h($primary) ?>">
                        <output><?= h($primary) ?></output>
                    </label>
                    <label class="color-field">
                        <span>Secondary color</span>
                        <input type="color" name="secondary_color" value="<?= h($secondary) ?>">
                        <output><?= h($secondary) ?></output>
                    </label>
                </div>

                <div class="theme-presets" aria-label="Theme presets">
                    <button type="button" data-primary="#15803D" data-secondary="#166534">Green</button>
                    <button type="button" data-primary="#2563EB" data-secondary="#1D4ED8">Blue</button>
                    <button type="button" data-primary="#7C3AED" data-secondary="#6D28D9">Purple</button>
                    <button type="button" data-primary="#C2410C" data-secondary="#9A3412">Orange</button>
                </div>
            </div>

            <div class="theme-preview" style="--preview-primary:<?= h($primary) ?>;--preview-secondary:<?= h($secondary) ?>">
                <span class="preview-dot"></span>
                <div><strong><?= h((string)($company['legal_name'] ?? 'Company')) ?></strong><span>Theme preview</span></div>
                <button type="button">Sample action</button>
            </div>

            <div class="form-actions">
                <button class="button" type="submit">Save theme</button>
            </div>
        </form>
        <?php endif; ?>
    </article>
</section>

<script>
(() => {
    const fields = [...document.querySelectorAll('.color-field input[type="color"]')];
    const preview = document.querySelector('.theme-preview');
    const sync = () => {
        if (!preview || fields.length < 2) return;
        preview.style.setProperty('--preview-primary', fields[0].value);
        preview.style.setProperty('--preview-secondary', fields[1].value);
        fields.forEach(field => field.parentElement.querySelector('output').textContent = field.value.toUpperCase());
    };
    fields.forEach(field => field.addEventListener('input', sync));
    document.querySelectorAll('.theme-presets button').forEach(button => button.addEventListener('click', () => {
        fields[0].value = button.dataset.primary;
        fields[1].value = button.dataset.secondary;
        sync();
    }));
})();
</script>
<?php else: ?>
<section class="empty card"><h2>Create a company first</h2><a class="button link-button" href="/companies/create">Create company</a></section>
<?php endif; ?>
