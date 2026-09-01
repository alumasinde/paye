<?php
$success=Session::flash('success');
$error=Session::flash('error');
$employees=(array)($run['employees']??[]);
$status=(string)($run['status']??'DRAFT');
$workflow=(array)($run['workflow']??[]);
$validation=(array)($validation??[]);$checks=(array)($validation['checks']??[]);$blocking=(int)($validation['blocking']??0);$warnings=(int)($validation['warnings']??0);

$steps=[
    'DRAFT'=>'Draft',
    'CALCULATED'=>'Calculated',
    'REVIEW'=>'Review',
    'APPROVED'=>'Approved',
    'FINALIZED'=>'Finalized',
    'LOCKED'=>'Locked',
];
$order=array_keys($steps);
$currentIndex=array_search($status,$order,true);
if ($status==='CALCULATION_FAILED') $currentIndex=0;
$currentIndex=$currentIndex===false?0:$currentIndex;

$actions=[
    'DRAFT'=>['calculate','Calculate payroll','Calculate all employee snapshots before review.',''],
    'CALCULATION_FAILED'=>['calculate','Retry calculation','Retry employees that could not be calculated. Resolve any employee salary or payroll rule issues shown below before sending the run for review.',''],
    'CALCULATED'=>['review','Send to review','Confirm the payroll figures and send this run into review.',''],
    'REVIEW'=>['approve','Approve payroll','Approve this reviewed payroll before finalization.','return confirm("Approve this payroll? Approved figures should only move forward to finalization.");'],
    'APPROVED'=>['finalize','Finalize payroll','Finalize the approved payroll and preserve the completed results.','return confirm("Finalize this payroll? This should only be done after the figures are fully approved.");'],
    'FINALIZED'=>['lock','Lock payroll','Lock the finalized payroll so its completed record cannot move through the workflow again.','return confirm("Lock this payroll? Treat this as the completed payroll record.");'],
];

$historyByStatus=[];
foreach($workflow as $event){
    $to=(string)($event['to_status']??'');
    if($to!=='') $historyByStatus[$to]=$event;
}
$statusClass='status-'.strtolower($status);
$editable=$status==='DRAFT' || $status==='CALCULATION_FAILED';
$calculated=0;$failed=0;$pending=0;$grossTotal=0.0;$netTotal=0.0;
foreach($employees as $employee){$es=(string)($employee['status']??'PENDING'); if($es==='CALCULATED'){$calculated++;$grossTotal+=(float)($employee['gross_salary']??0);$netTotal+=(float)($employee['net_salary']??0);} elseif($es==='FAILED'){$failed++;} else {$pending++;}}
?>
<?php if($success): ?><div class="alert success"><?= h($success) ?></div><?php endif; ?>
<?php if($error): ?><div class="alert error"><?= h($error) ?></div><?php endif; ?>

<section class="page-heading">
    <div>
        <a class="back-link" href="/payroll?company=<?= h($companyId) ?>">← Back to payroll</a>
        <p class="eyebrow">PAYROLL RUN · <?= h($status) ?></p>
        <h2><?= h((string)($run['period']??'')) ?> payroll</h2>
        <p class="muted"><?= count($employees) ?> employee<?= count($employees)===1?'':'s' ?> captured in this payroll run.</p>
    </div>
    <span class="status <?= h($statusClass) ?>"><?= h($status) ?></span>
</section>

<section class="card payroll-run-summary">
    <div class="section-heading"><div><p class="eyebrow">PAYROLL WORKSPACE</p><h2>At a glance</h2><p class="muted">Use the summary to identify what needs attention before moving this payroll forward.</p></div></div>
    <div class="stats-grid">
        <div class="stat-card"><span class="muted">Employees</span><strong><?= count($employees) ?></strong><small><?= $calculated ?> calculated</small></div>
        <div class="stat-card"><span class="muted">Pending</span><strong><?= $pending ?></strong><small><?= $failed ?> need attention</small></div>
        <div class="stat-card"><span class="muted">Gross pay</span><strong>KES <?= h(number_format($grossTotal,2)) ?></strong><small>Calculated employees</small></div>
        <div class="stat-card"><span class="muted">Net pay</span><strong>KES <?= h(number_format($netTotal,2)) ?></strong><small>Calculated employees</small></div>
    </div>
    <?php if($status==='CALCULATED' || $status==='CALCULATION_FAILED'): ?>
    <div class="workflow-action"><div><h3>Reopen payroll</h3><p class="muted">Return this calculated payroll to draft, clear calculation results, then refresh employees and recalculate. Adjustments are preserved.</p></div><form method="post" action="/payroll/action" onsubmit="return confirm('Reopen this payroll and clear all calculation results?');"><input type="hidden" name="_csrf" value="<?= h(csrf_token()) ?>"><input type="hidden" name="company_id" value="<?= h($companyId) ?>"><input type="hidden" name="run_id" value="<?= h((string)($run['id']??'')) ?>"><input type="hidden" name="action" value="reopen"><button class="button secondary" type="submit">Reopen to draft</button></form></div>
    <?php endif; ?>
    <?php if($editable): ?>
    <div class="workflow-action">
        <div><h3>Employee snapshot</h3><p class="muted">Refresh this draft after correcting employment dates or salary effective dates. Existing employee snapshots and adjustments are preserved.</p></div>
        <form method="post" action="/payroll/refresh-employees"><input type="hidden" name="_csrf" value="<?= h(csrf_token()) ?>"><input type="hidden" name="company_id" value="<?= h($companyId) ?>"><input type="hidden" name="run_id" value="<?= h((string)($run['id']??'')) ?>"><button class="button secondary" type="submit">Refresh employees</button></form>
    </div>
    <?php endif; ?>
</section>

<section id="validation" class="card">
 <div class="section-heading"><div><p class="eyebrow">PHASE 2 · VALIDATION</p><h2>Payroll validation & exceptions</h2><p class="muted"><?= $blocking ?> blocking issue<?= $blocking===1?'':'s' ?> · <?= $warnings ?> warning<?= $warnings===1?'':'s' ?></p></div></div>
 <?php if($checks): ?><div class="table-wrap"><table><thead><tr><th>Severity</th><th>Employee</th><th>Issue</th></tr></thead><tbody><?php foreach($checks as $check): ?><tr><td><span class="status small"><?= h((string)($check['severity']??'')) ?></span></td><td><?= h((string)($check['employee_name']??'Payroll')) ?></td><td><?= h((string)($check['message']??'')) ?></td></tr><?php endforeach; ?></tbody></table></div><?php else: ?><p class="muted">No validation issues found.</p><?php endif; ?>
</section>

<section class="workflow card">
    <div class="section-heading">
        <div>
            <h2>Payroll workflow</h2>
            <p class="muted">Each stage is recorded. Completed payrolls can be finalized and locked without recalculating historical results.</p>
        </div>
    </div>

    <div class="steps" aria-label="Payroll workflow">
        <?php foreach($steps as $stepStatus=>$stepLabel):
            $stepIndex=array_search($stepStatus,$order,true);
            $isCurrent=$stepIndex===$currentIndex;
            $isCompleted=$stepIndex<$currentIndex;
            $class='step'.($isCurrent?' current':($isCompleted?' completed':''));
            $event=$historyByStatus[$stepStatus]??null;
            $actor=$event!==null ? trim((string)($event['actor_name']??'')) : '';
        ?>
            <div class="<?= h($class) ?>">
                <span><?= h($stepLabel) ?></span>
                <?php if($actor!==''): ?>
                    <span class="step-meta"><?= h($actor) ?></span>
                <?php endif; ?>
            </div>
        <?php endforeach; ?>
    </div>

    <?php if(isset($actions[$status])): $next=$actions[$status]; ?>
        <div class="workflow-action">
            <div>
                <h3><?= h($next[1]) ?></h3>
                <p class="muted"><?= h($next[2]) ?></p>
            </div>
            <form method="post" action="/payroll/action" onsubmit="<?= h($next[3]) ?>">
                <input type="hidden" name="_csrf" value="<?= h(csrf_token()) ?>">
                <input type="hidden" name="company_id" value="<?= h($companyId) ?>">
                <input type="hidden" name="run_id" value="<?= h((string)($run['id']??'')) ?>">
                <input type="hidden" name="action" value="<?= h($next[0]) ?>">
                <button class="button" type="submit"><?= h($next[1]) ?></button>
            </form>
        </div>
    <?php else: ?>
        <div class="workflow-action">
            <div>
                <h3>Payroll locked</h3>
                <p class="muted">This payroll has completed its workflow and remains available for reporting and historical review.</p>
            </div>
        </div>
    <?php endif; ?>
</section>

<section class="card">
    <div class="section-heading">
        <div>
            <h2>Reports</h2>
            <p class="muted">Review payroll totals or open employee payslips.</p>
        </div>
        <a class="button secondary link-button" href="/reports/payroll?company=<?= rawurlencode($companyId) ?>&run=<?= rawurlencode((string)($run['id']??'')) ?>">View payroll summary</a>
    </div>
</section>

<section id="adjustments" class="card">
    <div class="section-heading">
        <div>
            <p class="eyebrow">PAYROLL BREAKDOWN</p>
            <h2>Earnings & adjustments</h2>
            <p class="muted">Add one-off earnings or deductions to this payroll run. Changes are stored with this payroll snapshot and are not written back to the employee's salary history.</p>
        </div>
        <?php if($status==='DRAFT' || $status==='CALCULATION_FAILED'): ?>
            <span class="status small status-draft">Editable</span>
        <?php else: ?>
            <span class="status small"><?= h($status) ?></span>
        <?php endif; ?>
    </div>

    <?php foreach($employees as $employee):
        $employeeName=trim((string)($employee['first_name']??'').' '.(string)($employee['middle_name']??'').' '.(string)($employee['last_name']??''));
        $adjustments=(array)($employee['adjustments']??[]);
        $earnings=0.0; $deductions=0.0;
        foreach($adjustments as $adjustment){
            if(($adjustment['kind']??'')==='EARNING') $earnings+=(float)($adjustment['amount']??0);
            if(($adjustment['kind']??'')==='DEDUCTION') $deductions+=(float)($adjustment['amount']??0);
        }
    ?>
        <details class="card" <?= count($employees)===1?'open':'' ?>>
            <summary>
                <strong><?= h($employeeName) ?></strong>
                <span class="muted"><?= h((string)($employee['employee_number']??'')) ?> · Earnings <?= h(number_format($earnings,2)) ?> · Deductions <?= h(number_format($deductions,2)) ?></span>
            </summary>
            <div class="section-heading">
                <div>
                    <h3>Current breakdown</h3>
                    <p class="muted">Basic salary is the payroll snapshot. Adjustments below apply only to this payroll period.</p>
                </div>
            </div>

            <div class="table-wrap">
                <table>
                    <thead><tr><th>Item</th><th>Type</th><th>Tax treatment</th><th>Amount</th><?php if($status==='DRAFT' || $status==='CALCULATION_FAILED'): ?><th></th><?php endif; ?></tr></thead>
                    <tbody>
                        <tr><td><strong>Basic salary</strong></td><td>BASE</td><td>Taxable</td><td><?= h(number_format((float)($employee['basic_salary']??0),2)) ?></td><?php if($status==='DRAFT' || $status==='CALCULATION_FAILED'): ?><td></td><?php endif; ?></tr>
                        <?php foreach($adjustments as $adjustment):
                            $kind=(string)($adjustment['kind']??'');
                            $treatment=$kind==='EARNING'
                                ? (!empty($adjustment['taxable']) ? 'Taxable' : 'Non-taxable')
                                : (!empty($adjustment['reduces_taxable_income']) ? 'Reduces taxable income' : 'Net-pay deduction');
                        ?>
                            <tr>
                                <td><?= h((string)($adjustment['name']??'')) ?></td>
                                <td><?= h($kind) ?></td>
                                <td><?= h($treatment) ?></td>
                                <td><?= h(number_format((float)($adjustment['amount']??0),2)) ?></td>
                                <?php if($status==='DRAFT' || $status==='CALCULATION_FAILED'): ?>
                                <td>
                                    <form method="post" action="/payroll/adjustment" onsubmit="return confirm('Remove this payroll adjustment?');">
                                        <input type="hidden" name="_csrf" value="<?= h(csrf_token()) ?>">
                                        <input type="hidden" name="company_id" value="<?= h($companyId) ?>">
                                        <input type="hidden" name="run_id" value="<?= h((string)($run['id']??'')) ?>">
                                        <input type="hidden" name="employee_run_id" value="<?= h((string)($employee['id']??'')) ?>">
                                        <input type="hidden" name="adjustment_id" value="<?= h((string)($adjustment['id']??'')) ?>">
                                        <input type="hidden" name="operation" value="delete">
                                        <button class="text-link" type="submit">Remove</button>
                                    </form>
                                </td>
                                <?php endif; ?>
                            </tr>
                        <?php endforeach; ?>
                    </tbody>
                </table>
            </div>

            <?php if($status==='DRAFT' || $status==='CALCULATION_FAILED'): ?>
            <form method="post" action="/payroll/adjustment" class="selector-card" style="margin-top:16px">
                <input type="hidden" name="_csrf" value="<?= h(csrf_token()) ?>">
                <input type="hidden" name="company_id" value="<?= h($companyId) ?>">
                <input type="hidden" name="run_id" value="<?= h((string)($run['id']??'')) ?>">
                <input type="hidden" name="employee_run_id" value="<?= h((string)($employee['id']??'')) ?>">
                <input type="hidden" name="operation" value="create">
                <div class="selectors">
                    <label>Item name<input name="name" maxlength="150" required placeholder="e.g. Transport allowance or Salary advance"></label>
                    <label>Type<select name="kind" required><option value="EARNING">Earning</option><option value="DEDUCTION">Deduction</option></select></label>
                    <label>Amount<input name="amount" inputmode="decimal" min="0.01" step="0.01" required placeholder="0.00"></label>
                    <label>Tax treatment
                        <span class="muted"><input type="checkbox" name="taxable" checked> Taxable earning</span>
                        <span class="muted"><input type="checkbox" name="reduces_taxable_income"> Deduction reduces taxable income</span>
                    </label>
                    <button class="button secondary" type="submit">Add adjustment</button>
                </div>
            </form>
            <?php endif; ?>
        </details>
    <?php endforeach; ?>
</section>

<section class="table-card card">
    <div class="section-heading">
        <div><p class="eyebrow">EMPLOYEES</p><h2>Employee payroll</h2><p class="muted">Search and filter the payroll population without losing the historical payroll snapshot.</p></div>
        <div class="selectors"><label>Search<input id="payrollEmployeeSearch" placeholder="Search name or employee number"></label><label>Status<select id="payrollEmployeeStatus"><option value="">All statuses</option><option value="PENDING">Pending</option><option value="CALCULATED">Calculated</option><option value="FAILED">Needs attention</option></select></label></div>
    </div>
    <div class="table-wrap">
        <table>
            <thead>
                <tr>
                    <th>Employee</th>
                    <th>Basic salary</th>
                    <th>Gross</th>
                    <th>PAYE</th>
                    <th>Deductions</th>
                    <th>Net pay</th>
                    <th>Status</th>
                </tr>
            </thead>
            <tbody>
                <?php foreach($employees as $employee):
                    $name=trim((string)($employee['first_name']??'').' '.(string)($employee['middle_name']??'').' '.(string)($employee['last_name']??''));
                    $employeeStatus=(string)($employee['status']??'PENDING');
                ?>
                    <tr class="payroll-employee-row" data-status="<?= h($employeeStatus) ?>" data-search="<?= h(strtolower($name.' '.(string)($employee['employee_number']??''))) ?>">
                        <td>
                            <strong><?= h($name) ?></strong>
                            <span class="muted"><?= h((string)($employee['employee_number']??'')) ?></span>
                        </td>
                        <td><?= h(number_format((float)($employee['basic_salary']??0),2)) ?></td>
                        <td><?= ($employee['gross_salary']??'')!=='' ? h(number_format((float)$employee['gross_salary'],2)) : 'Not calculated' ?></td>
                        <td><?= ($employee['paye']??'')!=='' ? h(number_format((float)$employee['paye'],2)) : 'Not calculated' ?></td>
                        <td><?= ($employee['total_deductions']??'')!=='' ? h(number_format((float)$employee['total_deductions'],2)) : 'Not calculated' ?></td>
                        <td><?= ($employee['net_salary']??'')!=='' ? h(number_format((float)$employee['net_salary'],2)) : 'Not calculated' ?></td>
                        <td>
                            <span class="status small status-<?= h(strtolower($employeeStatus)) ?>"><?= h($employeeStatus) ?></span>
                            <a class="text-link" href="/reports/payslip?company=<?= rawurlencode($companyId) ?>&run=<?= rawurlencode((string)($run['id']??'')) ?>&employee=<?= rawurlencode((string)($employee['id']??'')) ?>">Payslip</a>
                            <?php if(!empty($employee['error_message'])): ?><span class="row-error"><?= h((string)$employee['error_message']) ?></span><?php endif; ?>
                        </td>
                    </tr>
                <?php endforeach; ?>
            </tbody>
        </table>
    </div>
</section>
<script>
(()=>{const search=document.getElementById('payrollEmployeeSearch'),status=document.getElementById('payrollEmployeeStatus');if(!search||!status)return;const apply=()=>{const q=search.value.toLowerCase().trim(),s=status.value;document.querySelectorAll('.payroll-employee-row').forEach(row=>{row.hidden=(q&&!row.dataset.search.includes(q))||(s&&row.dataset.status!==s);});};search.addEventListener('input',apply);status.addEventListener('change',apply);})();
</script>
