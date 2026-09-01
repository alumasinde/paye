<?php
declare(strict_types=1);

final class AuthController
{
    public function showLogin(): void { view('auth/login', ['title' => 'Log in']); }
    public function showRegister(): void { view('auth/register', ['title' => 'Create account']); }

    public function login(): void
    {
        verify_csrf();
        $email = trim((string)($_POST['email'] ?? ''));
        $password = (string)($_POST['password'] ?? '');
        if (!filter_var($email, FILTER_VALIDATE_EMAIL) || $password === '') {
            Session::flash('error', 'Enter a valid email address and password.');
            redirect('/login');
        }
        $result = (new AuthService())->login($email, $password);
        if (!$result['ok']) { Session::flash('error', $result['message']); redirect('/login'); }
        $this->storeAuth($result['data']);
        redirect('/dashboard');
    }

    public function register(): void
    {
        verify_csrf();
        $input = [
            'first_name' => trim((string)($_POST['first_name'] ?? '')),
            'last_name' => trim((string)($_POST['last_name'] ?? '')),
            'email' => trim((string)($_POST['email'] ?? '')),
            'password' => (string)($_POST['password'] ?? ''),
        ];
        $confirm = (string)($_POST['confirm_password'] ?? '');
        if ($input['first_name'] === '' || $input['last_name'] === '' || !filter_var($input['email'], FILTER_VALIDATE_EMAIL)) {
            Session::flash('error', 'Complete all required fields with a valid email address.');
            redirect('/register');
        }
        if (strlen($input['password']) < 8) { Session::flash('error', 'Use a password with at least 8 characters.'); redirect('/register'); }
        if (!hash_equals($input['password'], $confirm)) { Session::flash('error', 'Passwords do not match.'); redirect('/register'); }

        $result = (new AuthService())->register($input);
        if (!$result['ok']) { Session::flash('error', $result['message']); redirect('/register'); }
        $this->storeAuth($result['data']);
        redirect('/dashboard');
    }

    public function logout(): void
    {
        verify_csrf();
        $_SESSION = [];
        if (ini_get('session.use_cookies')) {
            $params = session_get_cookie_params();
            setcookie(session_name(), '', time() - 42000, $params['path'], $params['domain'], (bool)$params['secure'], (bool)$params['httponly']);
        }
        session_destroy();
        redirect('/');
    }

    private function storeAuth(array $data): void
    {
        $tokens = (array)($data['tokens'] ?? []);
        $access = (string)($tokens['access_token'] ?? '');
        if ($access === '') { Session::flash('error', 'Authentication response did not include an access token.'); redirect('/login'); }
        Session::regenerate();
        $_SESSION['auth'] = [
            'user' => (array)($data['user'] ?? []),
            'access_token' => $access,
            'refresh_token' => (string)($tokens['refresh_token'] ?? ''),
        ];
    }
}
