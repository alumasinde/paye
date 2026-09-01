<?php
$app->router->get('/', [HomeController::class, 'index'], ['require_guest']);
$app->router->get('/login', [AuthController::class, 'showLogin'], ['require_guest']);
$app->router->post('/login', [AuthController::class, 'login'], ['require_guest']);
$app->router->get('/register', [AuthController::class, 'showRegister'], ['require_guest']);
$app->router->post('/register', [AuthController::class, 'register'], ['require_guest']);
$app->router->post('/logout', [AuthController::class, 'logout'], ['require_auth']);
$app->router->get('/dashboard', [DashboardController::class, 'index'], ['require_auth']);
$app->router->get('/companies', [CompanyController::class, 'index'], ['require_auth']);
$app->router->get('/companies/create', [CompanyController::class, 'showCreate'], ['require_auth']);
$app->router->post('/companies', [CompanyController::class, 'create'], ['require_auth']);

$app->router->get('/employees', [EmployeeController::class, 'index'], ['require_auth']);
$app->router->get('/employees/create', [EmployeeController::class, 'showCreate'], ['require_auth']);
$app->router->post('/employees', [EmployeeController::class, 'create'], ['require_auth']);
