<?php
declare(strict_types=1);

final class PayrollReportController
{
    private function companies(): array
    {
        $result = (new CompanyService())->list(auth_access_token());
        return $result['ok'] ? $result['companies'] : [];
    }

    public function summary(): void
    {
        $companyId = trim((string)($_GET['company'] ?? ''));
        $runId = trim((string)($_GET['run'] ?? ''));
        if ($companyId === '' || $runId === '') redirect('/payroll');

        $result = (new PayrollReportService())->get($companyId, $runId, auth_access_token());
        if (!$result['ok']) { Session::flash('error', $result['message']); redirect('/payroll?company=' . rawurlencode($companyId)); }

        view('reports/payroll-summary', ['title'=>'Payroll summary','layout'=>'app','activeNav'=>'reports','user'=>auth_user(),'companyId'=>$companyId,'report'=>$result['report']]);
    }

    public function payslip(): void
    {
        $companyId = trim((string)($_GET['company'] ?? ''));
        $runId = trim((string)($_GET['run'] ?? ''));
        $employeeId = trim((string)($_GET['employee'] ?? ''));
        if ($companyId === '' || $runId === '' || $employeeId === '') redirect('/payroll');

        $result = (new PayrollReportService())->get($companyId, $runId, auth_access_token());
        if (!$result['ok']) { Session::flash('error', $result['message']); redirect('/payroll?company=' . rawurlencode($companyId)); }

        $employee = null;
        foreach ($result['report']['employees'] as $item) if ((string)($item['id'] ?? '') === $employeeId) { $employee = $item; break; }
        if (!$employee) { Session::flash('error', 'Employee was not found in this payroll run.'); redirect('/reports/payroll?company=' . rawurlencode($companyId) . '&run=' . rawurlencode($runId)); }

        view('reports/payslip', ['title'=>'Payslip','layout'=>'app','activeNav'=>'reports','user'=>auth_user(),'companyId'=>$companyId,'run'=>$result['report']['run'],'employee'=>$employee]);
    }
}
