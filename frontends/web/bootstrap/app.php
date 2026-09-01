<?php
declare(strict_types=1);

require_once __DIR__ . '/../app/Support/functions.php';
require_once __DIR__ . '/../app/Core/Router.php';
require_once __DIR__ . '/../app/Core/App.php';
require_once __DIR__ . '/../app/Core/Session.php';
require_once __DIR__ . '/../app/Services/ApiService.php';
require_once __DIR__ . '/../app/Services/AuthService.php';
require_once __DIR__ . '/../app/Controllers/HomeController.php';
require_once __DIR__ . '/../app/Controllers/AuthController.php';
require_once __DIR__ . '/../app/Controllers/DashboardController.php';

$app = new App();
require __DIR__ . '/../routes/web.php';
return $app;
