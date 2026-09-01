<main class="center-page"><section class="auth-card">
<a class="public-brand compact" href="/"><span class="brand-mark">B</span> Budget254 Payroll</a>
<p class="eyebrow">WELCOME BACK</p><h1>Log in to your workspace</h1><p class="muted">Enter your account details to continue.</p>
<?php if ($error = Session::flash('error')): ?><div class="alert error"><?= h($error) ?></div><?php endif; ?>
<form method="post" class="form-stack" novalidate>
<input type="hidden" name="_csrf" value="<?= h(csrf_token()) ?>">
<label>Email address<input type="email" name="email" autocomplete="email" required></label>
<label>Password<input type="password" name="password" autocomplete="current-password" required></label>
<button class="button full" type="submit">Log in</button>
</form>
<p class="form-footer">New to Budget254 Payroll? <a href="/register">Create an account</a></p>
</section></main>