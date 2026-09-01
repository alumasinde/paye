<?php $error=Session::flash('error'); $old=$old??[]; $v=fn(string $key)=> (string)($old[$key]??$employee[$key]??''); ?>
<section class="page-heading">
  <div><a class="back-link" href="/employees/show?company=<?= h($companyId) ?>&employee=<?= h($employeeId) ?>">← Employee profile</a><p class="eyebrow">EMPLOYEE MANAGEMENT</p><h2>Edit employee</h2><p class="muted">Update employee information without changing historical payroll results.</p></div>
</section>
<?php if($error): ?><div class="alert error"><?= h($error) ?></div><?php endif; ?>
<form method="post" action="/employees/update" class="setup-card card">
<input type="hidden" name="_csrf" value="<?= h(csrf_token()) ?>"><input type="hidden" name="company_id" value="<?= h($companyId) ?>"><input type="hidden" name="employee_id" value="<?= h($employeeId) ?>">
<section class="form-section"><div><p class="eyebrow">01 · IDENTITY</p><h3>Personal record</h3></div><div class="form-grid">
<label>First name<input name="first_name" value="<?= h($v('first_name')) ?>" required></label><label>Middle name<input name="middle_name" value="<?= h($v('middle_name')) ?>"></label><label>Last name<input name="last_name" value="<?= h($v('last_name')) ?>" required></label>
<label>Gender<select name="gender"><option value="">Not specified</option><?php foreach(['MALE'=>'Male','FEMALE'=>'Female','OTHER'=>'Other','PREFER_NOT_TO_SAY'=>'Prefer not to say'] as $key=>$label): ?><option value="<?= $key ?>" <?= $v('gender')===$key?'selected':'' ?>><?= $label ?></option><?php endforeach; ?></select></label>
<label>Date of birth<input type="date" name="date_of_birth" value="<?= h(substr($v('date_of_birth'),0,10)) ?>"></label><label>Nationality<input name="nationality" value="<?= h($v('nationality')) ?>"></label><label>National ID<input name="national_id" value="<?= h($v('national_id')) ?>"></label><label>Passport number<input name="passport_number" value="<?= h($v('passport_number')) ?>"></label>
</div></section>
<section class="form-section"><div><p class="eyebrow">02 · EMPLOYMENT</p><h3>Work assignment and status</h3></div><div class="form-grid">
<label>Employment date<input type="date" name="employment_date" value="<?= h(substr($v('employment_date'),0,10)) ?>" required></label>
<label>Department<select name="department_id"><option value="">No department</option><?php foreach($departments as $department): $id=(string)($department['id']??''); ?><option value="<?= h($id) ?>" <?= $v('department_id')===$id?'selected':'' ?>><?= h((string)($department['name']??'Department')) ?></option><?php endforeach; ?></select></label>
<label>Job title<input name="job_title" value="<?= h($v('job_title')) ?>"></label>
<label>Employment type<select name="employment_type"><?php foreach(['PERMANENT','CONTRACT','CASUAL','INTERN','PART_TIME'] as $key): ?><option value="<?= $key ?>" <?= strtoupper($v('employment_type'))===$key?'selected':'' ?>><?= ucwords(strtolower(str_replace('_',' ',$key))) ?></option><?php endforeach; ?></select></label>
<label>Status<select name="status"><?php foreach(['ACTIVE','INACTIVE','SUSPENDED','TERMINATED'] as $key): ?><option value="<?= $key ?>" <?= strtoupper($v('status'))===$key?'selected':'' ?>><?= ucwords(strtolower($key)) ?></option><?php endforeach; ?></select></label>
<label>Termination date <span class="optional">If applicable</span><input type="date" name="termination_date" value="<?= h(substr($v('termination_date'),0,10)) ?>"></label>
</div></section>
<section class="form-section"><div><p class="eyebrow">03 · PAYROLL IDENTIFIERS</p><h3>Statutory details</h3></div><div class="form-grid">
<label>KRA PIN<input name="kra_pin" value="<?= h($v('kra_pin')) ?>"></label><label>NSSF number<input name="nssf_number" value="<?= h($v('nssf_number')) ?>"></label><label>SHIF number<input name="shif_number" value="<?= h($v('shif_number')) ?>"></label><label>NHIF number<input name="nhif_number" value="<?= h($v('nhif_number')) ?>"></label>
</div></section>
<div class="form-actions"><a class="button secondary link-button" href="/employees/show?company=<?= h($companyId) ?>&employee=<?= h($employeeId) ?>">Cancel</a><button class="button" type="submit">Save changes</button></div>
</form>