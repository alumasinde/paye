<?php
$success=Session::flash('success');
$error=Session::flash('error');
$employees=(array)($run['employees']??[]);
$status=(string)($run['status']??'DRAFT');
$workflow=(array)($run['workflow']??[]);

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
$currentIndex=$currentIndex===false?0:$currentIndex;

$actions=[
    'DRAFT'=>['calculate','Calculate payroll','Calculate all employee snapshots before review.',''],
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
            $class=$stepIndex===$currentIndex?'step current':'step';
        ?>
            <div class="<?= h($class) ?>">
                <?= h($stepLabel) ?>
                <?php if(isset($historyByStatus[$stepStatus])): ?>
                    <br><span><?= h((string)($historyByStatus[$stepStatus]['actor_name']??'')) ?></span>
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

<section class="table-card card">
    <div class="section-heading">
        <div>
            <h2>Employee payroll</h2>
            <p class="muted">Gross pay, deductions and net pay are stored against this payroll snapshot.</p>
        </div>
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
                    <tr>
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