<?php
declare(strict_types=1);
final class HomeController { public function index(): void { view('home/index', ['title' => 'Payroll for Kenyan businesses']); } }
