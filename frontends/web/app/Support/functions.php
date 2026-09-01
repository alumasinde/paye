<?php
declare(strict_types=1);
function h(string $value): string { return htmlspecialchars($value, ENT_QUOTES, 'UTF-8'); }
function base_path(string $path = ''): string { return __DIR__ . '/../..' . ($path === '' ? '' : '/' . ltrim($path, '/')); }
function view(string $name, array $data = []): void
{
    extract($data, EXTR_SKIP);
    $viewFile = base_path('app/Views/' . $name . '.php');
    if (!is_file($viewFile)) throw new RuntimeException('View not found.');
    $layout = ($data['layout'] ?? 'guest') === 'app' ? 'app.php' : 'guest.php';
    require base_path('app/Views/layouts/' . $layout);
}
function redirect(string $path): never { header('Location: ' . $path, true, 302); exit; }
function csrf_token(): string { if (empty($_SESSION['csrf_token'])) $_SESSION['csrf_token'] = bin2hex(random_bytes(32)); return (string)$_SESSION['csrf_token']; }
function verify_csrf(): void { $token=(string)($_POST['_csrf']??''); if($token===''||!hash_equals((string)($_SESSION['csrf_token']??''),$token)){http_response_code(419);exit('Your form session expired. Please go back and try again.');} }
function is_authenticated(): bool { return !empty($_SESSION['auth']['access_token']); }
function require_auth(): void { if (!is_authenticated()) redirect('/login'); }
function require_guest(): void { if (is_authenticated()) redirect('/dashboard'); }
function auth_user(): array { return (array)($_SESSION['auth']['user'] ?? []); }
function auth_access_token(): string { return (string)($_SESSION['auth']['access_token'] ?? ''); }


function company_context(): array
{
    static $loaded = false;
    static $context = [];
    if ($loaded) return $context;
    $loaded = true;

    $companyId = trim((string)($_GET['company'] ?? $_POST['company_id'] ?? ''));
    $context = [
        'id' => $companyId,
        'name' => '',
        'primary_color' => '#15803D',
        'secondary_color' => '#166534',
    ];

    if ($companyId === '' || !is_authenticated() || !class_exists('CompanyService')) return $context;

    try {
        $result = (new CompanyService())->get($companyId, auth_access_token());
        if ($result['ok']) {
            $company = (array)$result['company'];
            $tradingName = trim((string)($company['trading_name'] ?? ''));
            $context['name'] = $tradingName !== '' ? $tradingName : (string)($company['legal_name'] ?? '');
            $primary = trim((string)($company['primary_color'] ?? ''));
            $secondary = trim((string)($company['secondary_color'] ?? ''));
            if ($primary !== '') $context['primary_color'] = $primary;
            if ($secondary !== '') $context['secondary_color'] = $secondary;
        }
    } catch (Throwable) {
        // Keep the default theme if company context cannot be loaded.
    }

    return $context;
}
