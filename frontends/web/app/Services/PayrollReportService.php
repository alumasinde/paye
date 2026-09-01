<?php
declare(strict_types=1);

final class PayrollReportService
{
    public function __construct(private readonly PayrollRunService $runs = new PayrollRunService()) {}

    public function get(string $companyId, string $runId, string $token): array
    {
        $result = $this->runs->get($companyId, $runId, $token);
        if (!$result['ok']) return $result;

        $run = (array)$result['data'];
        $employees = (array)($run['employees'] ?? []);
        $totals = ['employees'=>count($employees),'basic_salary'=>0.0,'gross_salary'=>0.0,'paye'=>0.0,'statutory_deductions'=>0.0,'custom_deductions'=>0.0,'total_deductions'=>0.0,'net_salary'=>0.0];
        foreach ($employees as $employee) foreach (array_keys($totals) as $key) {
            if ($key === 'employees') continue;
            $totals[$key] += (float)($employee[$key] ?? 0);
        }
        $result['report'] = ['run'=>$run,'employees'=>$employees,'totals'=>$totals];
        return $result;
    }
}
