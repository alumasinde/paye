<?php
declare(strict_types=1);

final class Router
{
    private array $routes = [];

    public function get(string $path, callable|array $handler, array $middleware = []): void
    {
        $this->add('GET', $path, $handler, $middleware);
    }

    public function post(string $path, callable|array $handler, array $middleware = []): void
    {
        $this->add('POST', $path, $handler, $middleware);
    }

    private function add(string $method, string $path, callable|array $handler, array $middleware): void
    {
        $this->routes[$method][$path] = ['handler' => $handler, 'middleware' => $middleware];
    }

    public function dispatch(string $method, string $uri): void
    {
        $path = parse_url($uri, PHP_URL_PATH) ?: '/';
        $path = rtrim($path, '/') ?: '/';
        $route = $this->routes[$method][$path] ?? null;

        if ($route === null) {
            http_response_code(404);
            view('errors/404', ['title' => 'Page not found']);
            return;
        }

        foreach ($route['middleware'] as $middleware) {
            if (is_string($middleware) && function_exists($middleware)) { $middleware(); continue; }
            if (is_callable($middleware)) { $middleware(); }
        }

        $handler = $route['handler'];
        if (is_array($handler)) {
            [$class, $action] = $handler;
            (new $class())->{$action}();
            return;
        }

        $handler();
    }
}
