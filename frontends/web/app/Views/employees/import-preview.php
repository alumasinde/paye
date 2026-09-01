<section class="page-heading">
  <div>
    <a class="back-link" href="/employees/import?company=<?= h($companyId) ?>">← Upload another file</a>
    <p class="eyebrow">IMPORT REVIEW</p>
    <h2>Review employees</h2>
    <p class="muted"><?= count($valid) ?> valid · <?= count($invalid) ?> requiring attention</p>
  </div>
</section>

<?php if($invalid): ?>
<section class="table-card card">
  <div class="section-heading"><div><h2>Rows not ready</h2><p class="muted">These rows will not be imported. Correct them in the spreadsheet and upload again if needed.</p></div></div>
  <div class="table-wrap"><table>
    <thead><tr><th>Row</th><th>Employee</th><th>Employee no.</th><th>Issue</th></tr></thead>
    <tbody>
      <?php foreach($invalid as $row): ?><tr>
        <td><?= h((string)($row['_row']??'')) ?></td>
        <td><?= h(trim((string)($row['first_name']??'').' '.(string)($row['last_name']??''))) ?></td>
        <td><?= h((string)($row['employee_number']??'')) ?></td>
        <td><?= h(implode('; ', (array)($row['_errors']??[]))) ?></td>
      </tr><?php endforeach; ?>
    </tbody>
  </table></div>
</section>
<?php endif; ?>

<section class="table-card card">
  <div class="section-heading"><div><h2>Ready to import</h2><p class="muted">Nothing is created until you confirm.</p></div></div>
  <?php if($valid): ?>
    <div class="table-wrap"><table>
      <thead><tr><th>Employee</th><th>Employee no.</th><th>Department</th><th>Employment</th><th>Basic salary</th></tr></thead>
      <tbody>
        <?php foreach($valid as $row): ?><tr>
          <td><strong><?= h(trim((string)($row['first_name']??'').' '.(string)($row['middle_name']??'').' '.(string)($row['last_name']??''))) ?></strong><span class="muted"><?= h((string)($row['job_title']??'—')) ?></span></td>
          <td><?= h((string)($row['employee_number']??'')) ?></td>
          <td><?= h((string)($row['department_code']??'—')) ?></td>
          <td><?= h((string)($row['employment_type']??'')) ?></td>
          <td><?= h(number_format((float)($row['basic_salary']??0),2)) ?></td>
        </tr><?php endforeach; ?>
      </tbody>
    </table></div>
    <form method="post" action="/employees/import/confirm" class="form-actions">
      <input type="hidden" name="_csrf" value="<?= h(csrf_token()) ?>">
      <a class="button secondary link-button" href="/employees/import?company=<?= h($companyId) ?>">Cancel</a>
      <button class="button" type="submit">Import <?= count($valid) ?> employee<?= count($valid)===1?'':'s' ?></button>
    </form>
  <?php else: ?>
    <section class="empty"><h3>No rows are ready</h3><p class="muted">Correct the issues above and upload the spreadsheet again.</p><a class="button link-button" href="/employees/import?company=<?= h($companyId) ?>">Back to import</a></section>
  <?php endif; ?>
</section>