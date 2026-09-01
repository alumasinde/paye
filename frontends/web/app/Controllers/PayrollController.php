<?php
declare(strict_types=1);

final class PayrollController
{
    private function companies(): array
    {
        $result = (new CompanyService())->list(auth_access_token());
        return $result['ok'] ? $result['companies'] : [];
    }

    private function selectedCompany(array $companies): string
    {
        $id = trim((string)($_GET['company'] ?? $_POST['company_id'] ?? ''));
        return $id !== '' ? $id : (string)($companies[0]['id'] ?? '');
    }

    public function index(): void
    {
        $companies = $this->companies();
        $companyId = $this->selectedCompany($companies);
        $result = $companyId !== '' ? (new PayrollRunService())->list($companyId, auth_access_token()) : ['ok' => true, 'payroll_runs' => []];

        view('payroll/index', [
            'title' => 'Payroll',
            'layout' => 'app',
            'activeNav' => 'payroll',
            'user' => auth_user(),
            'companies' => $companies,
            'companyId' => $companyId,
            'runs' => $result['ok'] ? $result['payroll_runs'] : [],
            'loadError' => $result['ok'] ? '' : $result['message'],
        ]);
    }

    public function showCreate(): void
    {
        $companies = $this->companies();
        view('payroll/create', [
            'title' => 'Create payroll',
            'layout' => 'app',
            'activeNav' => 'payroll',
            'user' => auth_user(),
            'companies' => $companies,
            'companyId' => $this->selectedCompany($companies),
        ]);
    }

    public function create(): void
    {
        verify_csrf();
        $companies = $this->companies();
        $companyId = $this->selectedCompany($companies);
        $period = trim((string)($_POST['period'] ?? ''));

        if ($companyId === '' || !preg_match('/^\\d{4}-\\d{2}$/', $period)) {
            Session::flash('error', 'Select a company and choose a valid payroll month.');
            redirect('/payroll/create?company=' . rawurlencode($companyId));
        }

        $result = (new PayrollRunService())->create($companyId, $period, auth_access_token());
        if (!$result['ok']) {
            Session::flash('error', $result['message']);
            redirect('/payroll/create?company=' . rawurlencode($companyId));
        }

        $runId = (string)($result['data']['id'] ?? '');
        Session::flash('success', 'Payroll draft created. Review the employee snapshot and calculate when ready.');
        redirect($runId !== '' ? '/payroll/run?run=' . rawurlencode($runId) . '&company=' . rawurlencode($companyId) : '/payroll?company=' . rawurlencode($companyId));
    }

    public function show(): void
    {
        $companies = $this->companies();
        $companyId = $this->selectedCompany($companies);
        $runId = trim((string)($_GET['run'] ?? ''));
        $result = ($companyId !== '' && $runId !== '') ? (new PayrollRunService())->get($companyId, $runId, auth_access_token()) : ['ok' => false, 'message' => 'Payroll run was not found.'];

        if (!$result['ok']) {
            Session::flash('error', $result['message']);
            redirect('/payroll?company=' . rawurlencode($companyId));
        }

        view('payroll/show', [
            'title' => 'Payroll run',
            'layout' => 'app',
            'activeNav' => 'payroll',
            'user' => auth_user(),
            'companies' => $companies,
            'companyId' => $companyId,
            'run' => $result['data'],
        ]);
    }

    public function action(): void
    {
        verify_csrf();
        $companies = $this->companies();
        $companyId = $this->selectedCompany($companies);
        $runId = trim((string)($_POST['run_id'] ?? ''));
        $action = trim((string)($_POST['action'] ?? ''));

        $result = (new PayrollRunService())->action($companyId, $runId, $action, auth_access_token());
        Session::flash($result['ok'] ? 'success' : 'error', $result['ok'] ? ucfirst($action) . ' completed successfully.' : $result['message']);
        redirect('/payroll/run?run=' . rawurlencode($runId) . '&company=' . rawurlencode($companyId));
    }
}
