<main class="center-page"><section class="auth-card">
<a class="public-brand compact" href="/"><span class="brand-mark">B</span> Budget254 Payroll</a>
<p class="eyebrow">CREATE YOUR ACCOUNT</p><h1>Start your payroll workspace</h1><p class="muted">Your account is created first. Company setup follows after sign up.</p>
<?php if ($error = Session::flash('error')): ?><div class="alert error"><?= h($error) ?></div><?php endif; ?>
<form method="post" class="form-stack" novalidate>
<input type="hidden" name="_csrf" value="<?= h(csrf_token()) ?>">
<div class="form-grid"><label>First name<input type="text" name="first_name" autocomplete="given-name" required></label><label>Last name<input type="text" name="last_name" autocomplete="family-name" required></label></div>
<label>Email address<input type="email" name="email" autocomplete="email" required></label>
<label>Password<input type="password" name="password" minlength="8" autocomplete="new-password" required></label>
<label>Confirm password<input type="password" name="confirm_password" minlength="8" autocomplete="new-password" required></label>
<button class="button full" type="submit">Create account</button>
</form>
<p class="form-footer">Already have an account? <a href="/login">Log in</a></p>
</section></main>