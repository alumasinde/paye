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

$app->router->get('/departments', [DepartmentController::class, 'index'], ['require_auth']);
$app->router->get('/departments/create', [DepartmentController::class, 'showCreate'], ['require_auth']);
$app->router->post('/departments', [DepartmentController::class, 'create'], ['require_auth']);

$app->router->get('/employees', [EmployeeController::class, 'index'], ['require_auth']);
$app->router->get('/employees/create', [EmployeeController::class, 'showCreate'], ['require_auth']);
$app->router->post('/employees', [EmployeeController::class, 'create'], ['require_auth']);
$app->router->get('/employees/show', [EmployeeController::class, 'show'], ['require_auth']);
$app->router->get('/employees/edit', [EmployeeController::class, 'showEdit'], ['require_auth']);
$app->router->post('/employees/update', [EmployeeController::class, 'update'], ['require_auth']);
$app->router->post('/employees/salary', [EmployeeController::class, 'addSalary'], ['require_auth']);
$app->router->get('/employees/import', [EmployeeController::class, 'showImport'], ['require_auth']);
$app->router->post('/employees/import/preview', [EmployeeController::class, 'previewImport'], ['require_auth']);
$app->router->post('/employees/import/confirm', [EmployeeController::class, 'confirmImport'], ['require_auth']);
$app->router->get('/employees/export', [EmployeeController::class, 'export'], ['require_auth']);
$app->router->get('/employees/import-template', [EmployeeController::class, 'downloadTemplate'], ['require_auth']);

$app->router->get('/payroll', [PayrollController::class, 'index'], ['require_auth']);
$app->router->get('/payroll/create', [PayrollController::class, 'showCreate'], ['require_auth']);
$app->router->post('/payroll', [PayrollController::class, 'create'], ['require_auth']);
$app->router->get('/payroll/run', [PayrollController::class, 'show'], ['require_auth']);
$app->router->post('/payroll/action', [PayrollController::class, 'action'], ['require_auth']);
$app->router->post('/payroll/adjustment', [PayrollController::class, 'adjustment'], ['require_auth']);

$app->router->get('/reports/payroll', [PayrollReportController::class, 'summary'], ['require_auth']);
$app->router->get('/reports/payslip', [PayrollReportController::class, 'payslip'], ['require_auth']);

$app->router->get('/settings', [SettingsController::class, 'index'], ['require_auth']);
$app->router->post('/settings/theme', [SettingsController::class, 'updateTheme'], ['require_auth']);
