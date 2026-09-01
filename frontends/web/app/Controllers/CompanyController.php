<?php
declare(strict_types=1);

final class CompanyController
{
    public function index(): void
    {
        $service = new CompanyService();
        $result = $service->list(auth_access_token());

        view('companies/index', [
            'title' => 'Companies',
            'user' => auth_user(),
            'companies' => $result['ok'] ? $result['companies'] : [],
            'loadError' => $result['ok'] ? '' : $result['message'],
        ]);
    }

    public function showCreate(): void
    {
        view('companies/create', [
            'title' => 'Create company',
            'user' => auth_user(),
            'old' => Session::flash('old') ?: [],
        ]);
    }

    public function create(): void
    {
        verify_csrf();

        $input = [
            'legal_name' => trim((string)($_POST['legal_name'] ?? '')),
            'trading_name' => trim((string)($_POST['trading_name'] ?? '')),
            'kra_pin' => strtoupper(trim((string)($_POST['kra_pin'] ?? ''))),
            'email' => trim((string)($_POST['email'] ?? '')),
            'phone' => trim((string)($_POST['phone'] ?? '')),
            'country_code' => strtoupper(trim((string)($_POST['country_code'] ?? 'KE'))),
            'currency_code' => strtoupper(trim((string)($_POST['currency_code'] ?? 'KES'))),
            'payroll_frequency' => strtoupper(trim((string)($_POST['payroll_frequency'] ?? 'MONTHLY'))),
        ];

        Session::flash('old', $input);

        if ($input['legal_name'] === '' || $input['kra_pin'] === '' || !filter_var($input['email'], FILTER_VALIDATE_EMAIL)) {
            Session::flash('error', 'Enter the legal company name, KRA PIN and a valid company email address.');
            redirect('/companies/create');
        }

        if (!preg_match('/^[A-Z0-9]{8,15}$/', $input['kra_pin'])) {
            Session::flash('error', 'Enter a valid KRA PIN using letters and numbers only.');
            redirect('/companies/create');
        }

        $result = (new CompanyService())->create($input, auth_access_token());

        if (!$result['ok']) {
            Session::flash('error', $result['message']);
            redirect('/companies/create');
        }

        Session::flash('success', 'Company created successfully. You can now add employees and prepare payroll.');
        redirect('/companies');
    }
}
