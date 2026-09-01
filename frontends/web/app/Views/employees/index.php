<?php $success=Session::flash('success'); ?>
<?php if($success): ?><div class="alert success"><?= h($success) ?></div><?php endif; ?>
<?php if($loadError): ?><div class="alert error"><?= h($loadError) ?></div><?php endif; ?>
<section class="page-heading">
  <div><p class="eyebrow">PEOPLE</p><h2>Employees</h2><p class="muted">Add and manage employees separately for each company before preparing payroll.</p></div>
  <?php if($companies): ?><a class="button link-button" href="/employees/create?company=<?= h($companyId) ?>">Add employee</a><?php endif; ?>
</section>
<?php if($companies): ?>
<form method="get" class="selector-card card">
 <div class="selectors"><label>Company<select name="company" onchange="this.form.submit()">
 <?php foreach($companies as $company): $id=(string)($company['id']??''); ?><option value="<?= h($id) ?>" <?= $id===$companyId?'selected':'' ?>><?= h((string)($company['legal_name']??'Company')) ?></option><?php endforeach; ?>
 </select></label></div>
</form>
<?php if($employees): ?>
<section class="table-card card"><div class="section-heading"><div><h2>Employee directory</h2><p class="muted"><?= count($employees) ?> employee<?= count($employees)===1?'':'s' ?> in this company.</p></div></div>
<div class="table-wrap"><table><thead><tr><th>Employee</th><th>Employee no.</th><th>Department</th><th>Employment</th><th>Basic salary</th><th>Status</th></tr></thead><tbody>
<?php foreach($employees as $employee): ?><tr><td><strong><?= h(trim((string)($employee['first_name']??'').' '.(string)($employee['middle_name']??'').' '.(string)($employee['last_name']??''))) ?></strong><span class="muted"><?= h((string)($employee['job_title']??'—')) ?></span></td><td><?= h((string)($employee['employee_number']??'—')) ?></td><td><?= h((string)($employee['department']??'—')) ?></td><td><?= h((string)($employee['employment_type']??'—')) ?></td><td><?= h(number_format((float)($employee['current_salary']['basic_salary']??0),2)) ?></td><td><span class="status status-approved"><?= h((string)($employee['status']??'ACTIVE')) ?></span></td></tr><?php endforeach; ?>
</tbody></table></div></section>
<?php else: ?><section class="empty card"><p class="eyebrow">NO EMPLOYEES YET</p><h2>Start with your first employee</h2><p class="muted">Employee details and salary are linked to the selected company and kept ready for payroll.</p><a class="button link-button" href="/employees/create?company=<?= h($companyId) ?>">Add employee</a></section><?php endif; ?>
<?php else: ?><section class="empty card"><p class="eyebrow">COMPANY REQUIRED</p><h2>Create a company first</h2><p class="muted">Employees belong to a company so payroll records remain separated and auditable.</p><a class="button link-button" href="/companies/create">Create company</a></section><?php endif; ?>
