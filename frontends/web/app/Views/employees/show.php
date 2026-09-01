<?php
$success=Session::flash('success'); $error=Session::flash('error');
$name=trim((string)($employee['first_name']??'').' '.(string)($employee['middle_name']??'').' '.(string)($employee['last_name']??''));
$salary=(array)($employee['current_salary']??[]);
$status=strtoupper((string)($employee['status']??'ACTIVE'));
?>
<?php if($success): ?><div class="alert success"><?= h($success) ?></div><?php endif; ?>
<?php if($error): ?><div class="alert error"><?= h($error) ?></div><?php endif; ?>

<section class="page-heading employee-profile-heading">
  <div>
    <a class="back-link" href="/employees?company=<?= h($companyId) ?>">← Employees</a>
    <p class="eyebrow">EMPLOYEE PROFILE</p>
    <h2><?= h($name!==''?$name:'Employee') ?></h2>
    <p class="muted"><?= h((string)($employee['employee_number']??'')) ?><?= !empty($employee['job_title'])?' · '.h((string)$employee['job_title']):'' ?></p>
  </div>
  <div class="page-actions">
    <span class="status status-<?= strtolower(h($status)) ?>"><?= h($status) ?></span>
    <a class="button link-button" href="/employees/edit?company=<?= h($companyId) ?>&employee=<?= h((string)($employee['id']??'')) ?>">Edit employee</a>
  </div>
</section>

<section class="profile-grid">
  <section class="card profile-card">
    <div class="section-heading"><div><p class="eyebrow">CURRENT PAY</p><h2>KES <?= h(number_format((float)($salary['basic_salary']??0),2)) ?></h2><p class="muted"><?= h((string)($salary['pay_frequency']??'—')) ?> · effective <?= h(substr((string)($salary['effective_from']??''),0,10)) ?></p></div></div>
  </section>
  <section class="card profile-card">
    <div class="section-heading"><div><p class="eyebrow">EMPLOYMENT</p><h2><?= h((string)($employee['department']??'Unassigned')) ?></h2><p class="muted"><?= h((string)($employee['employment_type']??'—')) ?> · since <?= h(substr((string)($employee['employment_date']??''),0,10)) ?></p></div></div>
  </section>
</section>

<section class="detail-grid">
  <section class="table-card card">
    <div class="section-heading"><div><h2>Employee details</h2><p class="muted">Identity and employment record.</p></div></div>
    <dl class="detail-list">
      <div><dt>Employee number</dt><dd><?= h((string)($employee['employee_number']??'—')) ?></dd></div>
      <div><dt>Department</dt><dd><?= h((string)($employee['department']??'Unassigned')) ?></dd></div>
      <div><dt>Job title</dt><dd><?= h((string)($employee['job_title']??'—')) ?></dd></div>
      <div><dt>Employment type</dt><dd><?= h((string)($employee['employment_type']??'—')) ?></dd></div>
      <div><dt>Employment date</dt><dd><?= h(substr((string)($employee['employment_date']??''),0,10)?:'—') ?></dd></div>
      <div><dt>Termination date</dt><dd><?= h(substr((string)($employee['termination_date']??''),0,10)?:'—') ?></dd></div>
      <div><dt>Gender</dt><dd><?= h((string)($employee['gender']??'—')) ?></dd></div>
      <div><dt>Date of birth</dt><dd><?= h(substr((string)($employee['date_of_birth']??''),0,10)?:'—') ?></dd></div>
    </dl>
  </section>

  <section class="table-card card">
    <div class="section-heading"><div><h2>Payroll identifiers</h2><p class="muted">Statutory information used by payroll.</p></div></div>
    <dl class="detail-list">
      <div><dt>KRA PIN</dt><dd><?= h((string)($employee['kra_pin']??'—')) ?></dd></div>
      <div><dt>NSSF number</dt><dd><?= h((string)($employee['nssf_number']??'—')) ?></dd></div>
      <div><dt>SHIF number</dt><dd><?= h((string)($employee['shif_number']??'—')) ?></dd></div>
      <div><dt>NHIF number</dt><dd><?= h((string)($employee['nhif_number']??'—')) ?></dd></div>
      <div><dt>National ID</dt><dd><?= h((string)($employee['national_id']??'—')) ?></dd></div>
      <div><dt>Passport number</dt><dd><?= h((string)($employee['passport_number']??'—')) ?></dd></div>
    </dl>
  </section>
</section>

<section class="table-card card">
  <div class="section-heading"><div><p class="eyebrow">PAY HISTORY</p><h2>Salary history</h2><p class="muted">Salary changes are effective-dated and preserve payroll history.</p></div></div>
  <?php if($historyError): ?><div class="alert error"><?= h($historyError) ?></div><?php endif; ?>
  <div class="table-wrap"><table>
    <thead><tr><th>Basic salary</th><th>Frequency</th><th>Effective from</th><th>Effective to</th></tr></thead>
    <tbody><?php foreach($salaryHistory as $row): ?><tr><td><strong>KES <?= h(number_format((float)($row['basic_salary']??0),2)) ?></strong></td><td><?= h((string)($row['pay_frequency']??'—')) ?></td><td><?= h(substr((string)($row['effective_from']??''),0,10)) ?></td><td><?= h(substr((string)($row['effective_to']??''),0,10)?:'Current') ?></td></tr><?php endforeach; ?></tbody>
  </table></div>
</section>

<section class="setup-card card">
  <div class="section-heading"><div><p class="eyebrow">NEW SALARY</p><h2>Schedule salary change</h2><p class="muted">The new salary starts on its effective date and the previous record closes automatically.</p></div></div>
  <form method="post" action="/employees/salary" class="compact-form">
    <input type="hidden" name="_csrf" value="<?= h(csrf_token()) ?>">
    <input type="hidden" name="company_id" value="<?= h($companyId) ?>">
    <input type="hidden" name="employee_id" value="<?= h((string)($employee['id']??'')) ?>">
    <label>Basic salary<input type="number" name="basic_salary" min="0" step="0.01" required></label>
    <label>Pay frequency<select name="pay_frequency"><option value="MONTHLY">Monthly</option><option value="WEEKLY">Weekly</option><option value="BIWEEKLY">Bi-weekly</option></select></label>
    <label>Effective from<input type="date" name="effective_from" required></label>
    <button class="button" type="submit">Add salary change</button>
  </form>
</section>