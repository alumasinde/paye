<?php
declare(strict_types=1);

final class PayrollRunService
{
    public function __construct(private readonly ApiService $api = new ApiService()) {}

    public function list(string $companyId, string $token): array
    {
        $result = $this->api->request('GET', '/companies/' . rawurlencode($companyId) . '/payroll-runs', null, $token);
        if ($result['ok']) $result['payroll_runs'] = (array)($result['data']['payroll_runs'] ?? []);
        return $result;
    }

    public function create(string $companyId, string $period, string $token): array
    {
        return $this->api->request('POST', '/companies/' . rawurlencode($companyId) . '/payroll-runs', ['period' => $period], $token);
    }

    public function get(string $companyId, string $runId, string $token): array
    {
        return $this->api->request('GET', '/companies/' . rawurlencode($companyId) . '/payroll-runs/' . rawurlencode($runId), null, $token);
    }

    public function action(string $companyId, string $runId, string $action, string $token): array
    {
        $allowed = ['calculate', 'review', 'approve', 'finalize', 'lock'];
        if (!in_array($action, $allowed, true)) return ['ok' => false, 'message' => 'Unsupported payroll action.'];
        return $this->api->request('POST', '/companies/' . rawurlencode($companyId) . '/payroll-runs/' . rawurlencode($runId) . '/' . $action, [], $token);
    }
}
