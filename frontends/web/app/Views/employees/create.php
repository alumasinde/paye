<?php $error=Session::flash('error'); $old=$old??[]; ?>
<div class="page-heading"><div><a class="back-link" href="/employees?company=<?= h($companyId) ?>">← Back to employees</a><p class="eyebrow">EMPLOYEE SETUP</p><h2>Add employee</h2><p class="muted">Capture the employee identity, statutory details and current salary in one guided form.</p></div></div>
<?php if($error): ?><div class="alert error"><?= h($error) ?></div><?php endif; ?>
<?php if(!$companies): ?><section class="empty card"><h2>Create a company first</h2><a class="button link-button" href="/companies/create">Create company</a></section><?php else: ?>
<form method="post" action="/employees" class="setup-card card" novalidate><input type="hidden" name="_csrf" value="<?= h(csrf_token()) ?>">
<section class="form-section"><div><p class="eyebrow">01 · COMPANY & IDENTITY</p><h3>Who is this employee?</h3></div><div class="form-grid">
<label>Company<select name="company_id" required><?php foreach($companies as $company): $id=(string)($company['id']??''); ?><option value="<?= h($id) ?>" <?= $id===$companyId?'selected':'' ?>><?= h((string)($company['legal_name']??'Company')) ?></option><?php endforeach; ?></select></label>
<label>Employee number<input name="employee_number" value="<?= h((string)($old['employee_number']??'')) ?>" required></label>
<label>First name<input name="first_name" value="<?= h((string)($old['first_name']??'')) ?>" required></label>
<label>Middle name <span class="optional">Optional</span><input name="middle_name" value="<?= h((string)($old['middle_name']??'')) ?>"></label>
<label>Last name<input name="last_name" value="<?= h((string)($old['last_name']??'')) ?>" required></label>
<label>Gender<select name="gender"><option value="">Not specified</option><?php foreach(['MALE'=>'Male','FEMALE'=>'Female','OTHER'=>'Other','PREFER_NOT_TO_SAY'=>'Prefer not to say'] as $value=>$label): ?><option value="<?= $value ?>" <?= ($old['gender']??'')===$value?'selected':'' ?>><?= $label ?></option><?php endforeach; ?></select></label>
</div></section>
<section class="form-section"><div><p class="eyebrow">02 · EMPLOYMENT</p><h3>Role and start details</h3></div><div class="form-grid">
<label>Employment date<input type="date" name="employment_date" value="<?= h((string)($old['employment_date']??date('Y-m-d'))) ?>" required></label>
<label>Job title<input name="job_title" value="<?= h((string)($old['job_title']??'')) ?>"></label>
<label>Department<select name="department_id"><option value="">No department</option><?php foreach(($departments??[]) as $department): $id=(string)($department['id']??''); ?><option value="<?= h($id) ?>" <?= $id===(string)($old['department_id']??'')?'selected':'' ?>><?= h((string)($department['name']??'Department')) ?><?= !empty($department['code'])?' · '.h((string)$department['code']):'' ?></option><?php endforeach; ?></select><span class="muted">Departments are specific to the selected company.</span></label>
<label>Employment type<select name="employment_type"><?php foreach(['PERMANENT','CONTRACT','CASUAL','INTERN','PART_TIME'] as $v): ?><option value="<?= $v ?>" <?= ($old['employment_type']??'PERMANENT')===$v?'selected':'' ?>><?= ucwords(strtolower(str_replace('_',' ',$v))) ?></option><?php endforeach; ?></select></label>
</div></section>
<section class="form-section"><div><p class="eyebrow">03 · STATUTORY & SALARY</p><h3>Payroll information</h3></div><div class="form-grid">
<label>KRA PIN<input name="kra_pin" value="<?= h((string)($old['kra_pin']??'')) ?>"></label><label>NSSF number<input name="nssf_number" value="<?= h((string)($old['nssf_number']??'')) ?>"></label><label>SHIF number<input name="shif_number" value="<?= h((string)($old['shif_number']??'')) ?>"></label><label>Basic monthly salary<input type="number" min="0" step="0.01" name="basic_salary" value="<?= h((string)($old['basic_salary']??'')) ?>" required></label>
<label>Pay frequency<select name="pay_frequency"><option value="MONTHLY">Monthly</option><option value="WEEKLY">Weekly</option><option value="BIWEEKLY">Bi-weekly</option></select></label><label>Salary effective from<input type="date" name="effective_from" value="<?= h((string)($old['effective_from']??date('Y-m-d'))) ?>" required></label>
</div></section>
<div class="form-actions"><a class="button secondary link-button" href="/employees?company=<?= h($companyId) ?>">Cancel</a><button class="button" type="submit">Save employee</button></div></form><?php endif; ?>
