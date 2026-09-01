<?php
declare(strict_types=1);

final class DashboardController
{
    public function index(): void
    {
        $user = auth_user();
        view('dashboard/index', ['title' => 'Dashboard', 'user' => $user]);
    }
}
