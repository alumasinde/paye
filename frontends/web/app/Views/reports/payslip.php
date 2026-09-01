<?php
declare(strict_types=1);

$name = trim(implode(' ', array_filter([
    trim((string)($employee['first_name'] ?? '')),
    trim((string)($employee['middle_name'] ?? '')),
    trim((string)($employee['last_name'] ?? '')),
])));
$company = (array)($company ?? []);
$companyName = trim((string)($company['trading_name'] ?? ''));
if ($companyName === '') $companyName = trim((string)($company['legal_name'] ?? ''));
$period = (string)($run['period'] ?? '');
$periodStart = trim((string)($run['period_start'] ?? ''));
$periodEnd = trim((string)($run['period_end'] ?? ''));
$periodLabel = $period;
if ($periodStart !== '' && strtotime($periodStart) !== false) $periodLabel = date('F Y', strtotime($periodStart));
$money = static fn(mixed $value): string => 'KES ' . number_format((float)$value, 2);
$amount = static fn(mixed $value): float => (float)($value ?? 0);

$adjustments = (array)($employee['adjustments'] ?? []);
$earningAdjustments = [];
$deductionAdjustments = [];
foreach ($adjustments as $adjustment) {
    $kind = strtoupper(trim((string)($adjustment['kind'] ?? '')));
    if ($kind === 'EARNING') $earningAdjustments[] = (array)$adjustment;
    if ($kind === 'DEDUCTION') $deductionAdjustments[] = (array)$adjustment;
}
$deductionAdjustmentTotal = array_sum(array_map(static fn(array $item): float => $amount($item['amount'] ?? 0), $deductionAdjustments));
$customDeductions = $amount($employee['custom_deductions'] ?? 0);
$otherCustomDeductions = max(0, $customDeductions - $deductionAdjustmentTotal);
$currency = strtoupper(trim((string)($company['currency_code'] ?? 'KES')));
if ($currency === '') $currency = 'KES';
?>
<section class="payslip-page">
    <div class="payslip-toolbar no-print">
        <a class="back-link" href="/reports/payroll?company=<?= rawurlencode($companyId) ?>&run=<?= rawurlencode((string)($run['id'] ?? '')) ?>">← Payroll summary</a>
        <div class="payslip-toolbar-actions">
            <button class="button secondary" type="button" onclick="window.print()">Print / Save PDF</button>
        </div>
    </div>

    <article class="payslip-document" aria-label="Employee payslip">
        <header class="payslip-document-header">
            <div>
                <h2 class="payslip-company-name"><?= h($companyName !== '' ? $companyName : 'Company') ?></h2>
                <div class="payslip-company-meta">
                    <?php if (trim((string)($company['legal_name'] ?? '')) !== '' && trim((string)($company['legal_name'] ?? '')) !== $companyName): ?><span><?= h((string)$company['legal_name']) ?></span><?php endif; ?>
                    <?php if (trim((string)($company['kra_pin'] ?? '')) !== ''): ?><span>KRA PIN: <?= h((string)$company['kra_pin']) ?></span><?php endif; ?>
                    <?php if (trim((string)($company['email'] ?? '')) !== ''): ?><span><?= h((string)$company['email']) ?></span><?php endif; ?>
                    <?php if (trim((string)($company['phone'] ?? '')) !== ''): ?><span><?= h((string)$company['phone']) ?></span><?php endif; ?>
                </div>
            </div>
            <div class="payslip-title-block">
                <p class="payslip-kicker">SALARY STATEMENT</p>
                <h1 class="payslip-title">PAYSLIP</h1>
                <p class="payslip-period"><?= h($periodLabel) ?></p>
            </div>
        </header>

        <section class="payslip-employee">
            <div><span class="payslip-field-label">Employee</span><div class="payslip-field-value"><?= h($name !== '' ? $name : 'Employee') ?></div></div>
            <div><span class="payslip-field-label">Employee number</span><div class="payslip-field-value"><?= h((string)($employee['employee_number'] ?? '—')) ?></div></div>
            <div><span class="payslip-field-label">Pay frequency</span><div class="payslip-field-value"><?= h(ucfirst(strtolower((string)($employee['pay_frequency'] ?? 'MONTHLY')))) ?></div></div>
        </section>

        <div class="payslip-body">
            <section class="payslip-section">
                <h2 class="payslip-section-title">Earnings</h2>
                <table class="payslip-table">
                    <thead><tr><th>Description</th><th>Amount (<?= h($currency) ?>)</th></tr></thead>
                    <tbody>
                        <tr><td>Basic salary</td><td><?= h($money($employee['basic_salary'] ?? 0)) ?></td></tr>
                        <?php foreach ($earningAdjustments as $adjustment): ?><tr><td><?= h((string)($adjustment['name'] ?? 'Additional earning')) ?></td><td><?= h($money($adjustment['amount'] ?? 0)) ?></td></tr><?php endforeach; ?>
                        <tr class="subtotal"><td>Gross salary</td><td><?= h($money($employee['gross_salary'] ?? 0)) ?></td></tr>
                    </tbody>
                </table>
            </section>

            <section class="payslip-section">
                <h2 class="payslip-section-title">Tax calculation</h2>
                <table class="payslip-table"><tbody>
                    <tr><td class="muted-line">Taxable income</td><td><?= h($money($employee['taxable_income'] ?? 0)) ?></td></tr>
                    <tr><td class="muted-line">PAYE before relief</td><td><?= h($money($employee['paye_before_relief'] ?? 0)) ?></td></tr>
                    <tr><td class="muted-line">Personal relief</td><td><?= h('-' . $money($employee['relief'] ?? 0)) ?></td></tr>
                    <tr class="subtotal"><td>PAYE</td><td><?= h($money($employee['paye'] ?? 0)) ?></td></tr>
                </tbody></table>
            </section>

            <section class="payslip-section">
                <h2 class="payslip-section-title">Deductions</h2>
                <table class="payslip-table">
                    <thead><tr><th>Description</th><th>Amount (<?= h($currency) ?>)</th></tr></thead>
                    <tbody>
                        <tr><td>PAYE</td><td><?= h($money($employee['paye'] ?? 0)) ?></td></tr>
                        <tr><td>Statutory deductions</td><td><?= h($money($employee['statutory_deductions'] ?? 0)) ?></td></tr>
                        <?php foreach ($deductionAdjustments as $adjustment): ?><tr><td><?= h((string)($adjustment['name'] ?? 'Other deduction')) ?></td><td><?= h($money($adjustment['amount'] ?? 0)) ?></td></tr><?php endforeach; ?>
                        <?php if ($otherCustomDeductions > 0.00001): ?><tr><td>Other deductions</td><td><?= h($money($otherCustomDeductions)) ?></td></tr><?php endif; ?>
                        <tr class="subtotal"><td>Total deductions</td><td><?= h($money($employee['total_deductions'] ?? 0)) ?></td></tr>
                    </tbody>
                </table>
            </section>

            <section class="payslip-summary">
                <div class="payslip-tax-note">
                    Payroll period: <strong><?= h($periodLabel) ?></strong><br>
                    Payroll status: <strong><?= h((string)($run['status'] ?? '')) ?></strong>
                    <?php if ($periodEnd !== '' && strtotime($periodEnd) !== false): ?><br>Period ending: <strong><?= h(date('d M Y', strtotime($periodEnd))) ?></strong><?php endif; ?>
                </div>
                <div class="payslip-totals">
                    <div class="payslip-total-row"><span>Gross pay</span><strong><?= h($money($employee['gross_salary'] ?? 0)) ?></strong></div>
                    <div class="payslip-total-row"><span>Total deductions</span><strong><?= h($money($employee['total_deductions'] ?? 0)) ?></strong></div>
                    <div class="payslip-net-pay"><span>Net pay</span><strong><?= h($money($employee['net_salary'] ?? 0)) ?></strong></div>
                </div>
            </section>
        </div>

        <footer class="payslip-footer">
            <span>Generated from the payroll snapshot for <?= h($periodLabel) ?>.</span>
            <span class="payslip-print-only">Employee copy</span>
        </footer>
    </article>
</section>
