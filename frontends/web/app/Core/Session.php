<?php
declare(strict_types=1);
final class Session
{
    public static function start(): void
    {
        if (session_status() === PHP_SESSION_ACTIVE) return;
        session_name('budget254_payroll');
        session_set_cookie_params(['httponly'=>true,'secure'=>(!empty($_SERVER['HTTPS']) && $_SERVER['HTTPS'] !== 'off'),'samesite'=>'Lax','path'=>'/']);
        session_start();
    }
    public static function regenerate(): void { session_regenerate_id(true); }
    public static function flash(string $key, mixed $value = null): mixed
    {
        if (func_num_args() > 1) { $_SESSION['_flash'][$key] = $value; return null; }
        $value = $_SESSION['_flash'][$key] ?? null;
        unset($_SESSION['_flash'][$key]);
        return $value;
    }
}
