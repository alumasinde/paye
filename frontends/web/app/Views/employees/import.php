<?php $error = Session::flash('error'); ?>
<section class="page-heading">
  <div>
    <a class="back-link" href="/employees?company=<?= h($companyId) ?>">← Back to employees</a>
    <p class="eyebrow">BULK IMPORT</p>
    <h2>Import employees</h2>
    <p class="muted">Upload a spreadsheet, review every row, then confirm the valid employees.</p>
  </div>
  <a class="button secondary link-button" href="/employees/import-template">Download template</a>
</section>

<?php if($error): ?><div class="alert error"><?= h($error) ?></div><?php endif; ?>

<section class="setup-card card">
  <form method="post" action="/employees/import/preview" enctype="multipart/form-data">
    <input type="hidden" name="_csrf" value="<?= h(csrf_token()) ?>">
    <section class="form-section">
      <div><p class="eyebrow">01 · COMPANY</p><h3>Choose where employees belong</h3></div>
      <div class="form-grid">
        <label>Company
          <select name="company_id" required>
            <option value="">Select company</option>
            <?php foreach($companies as $company): $id=(string)($company['id']??''); ?>
              <option value="<?= h($id) ?>" <?= $id===$companyId?'selected':'' ?>><?= h((string)($company['legal_name']??'Company')) ?></option>
            <?php endforeach; ?>
          </select>
        </label>
      </div>
    </section>

    <section class="form-section">
      <div><p class="eyebrow">02 · SPREADSHEET</p><h3>Upload employee data</h3></div>
      <div class="form-grid">
        <label>CSV or XLSX file
          <input type="file" name="spreadsheet" accept=".csv,.xlsx,text/csv,application/vnd.openxmlformats-officedocument.spreadsheetml.sheet" required>
          <span class="muted">Maximum 5 MB · up to 1,000 employees</span>
        </label>
      </div>
    </section>

    <div class="form-actions">
      <a class="button secondary link-button" href="/employees?company=<?= h($companyId) ?>">Cancel</a>
      <button class="button" type="submit">Validate and review</button>
    </div>
  </form>
</section>

<section class="table-card card">
  <div class="section-heading"><div><h2>Import rules</h2><p class="muted">Required columns are checked before anything is created.</p></div></div>
  <div class="table-wrap"><table>
    <thead><tr><th>Required</th><th>Optional</th></tr></thead>
    <tbody><tr><td><strong>employee_number · first_name · last_name · employment_date · basic_salary</strong></td><td>middle_name · gender · job_title · department_code · statutory numbers · employment type · pay frequency</td></tr></tbody>
  </table></div>
</section>