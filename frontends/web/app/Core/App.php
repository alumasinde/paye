<?php
declare(strict_types=1);

final class App
{
    public Router $router;

    public function __construct()
    {
        Session::start();
        $this->router = new Router();
    }

    public function run(): void
    {
        $this->router->dispatch($_SERVER['REQUEST_METHOD'] ?? 'GET', $_SERVER['REQUEST_URI'] ?? '/');
    }
}
