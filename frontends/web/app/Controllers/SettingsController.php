<?php
declare(strict_types=1);

final class SettingsController
{
    public function index(): void
    {
        $service = new CompanyService();
        $result = $service->list(auth_access_token());
        $companies = $result['ok'] ? $result['companies'] : [];
        $companyId = trim((string)($_GET['company'] ?? ''));
        if ($companyId === '' && $companies) $companyId = (string)($companies[0]['id'] ?? '');

        $company = [];
        $loadError = '';
        if ($companyId !== '') {
            $detail = $service->get($companyId, auth_access_token());
            if ($detail['ok']) $company = $detail['company'];
            else $loadError = $detail['message'];
        }

        view('settings/index', [
            'title' => 'Settings',
            'layout' => 'app',
            'activeNav' => 'settings',
            'user' => auth_user(),
            'companies' => $companies,
            'companyId' => $companyId,
            'company' => $company,
            'loadError' => $loadError,
        ]);
    }

    public function updateTheme(): void
    {
        verify_csrf();
        $companyId = trim((string)($_POST['company_id'] ?? ''));
        $primary = strtoupper(trim((string)($_POST['primary_color'] ?? '')));
        $secondary = strtoupper(trim((string)($_POST['secondary_color'] ?? '')));

        if ($companyId === '' || !preg_match('/^#[0-9A-F]{6}$/', $primary) || !preg_match('/^#[0-9A-F]{6}$/', $secondary)) {
            Session::flash('error', 'Choose valid primary and secondary colors.');
            redirect('/settings?company=' . rawurlencode($companyId));
        }

        $service = new CompanyService();
        $detail = $service->get($companyId, auth_access_token());
        if (!$detail['ok']) {
            Session::flash('error', $detail['message']);
            redirect('/settings');
        }

        $company = $detail['company'];
        $company['primary_color'] = $primary;
        $company['secondary_color'] = $secondary;

        $result = $service->update($companyId, $company, auth_access_token());
        if (!$result['ok']) {
            Session::flash('error', $result['message']);
            redirect('/settings?company=' . rawurlencode($companyId));
        }

        Session::flash('success', 'Company theme updated.');
        redirect('/settings?company=' . rawurlencode($companyId));
    }
}
