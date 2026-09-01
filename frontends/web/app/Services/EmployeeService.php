<?php
declare(strict_types=1);
final class EmployeeService
{
    public function __construct(private readonly ApiService $api = new ApiService()) {}
    public function list(string $companyId,string $token): array {
        $result=$this->api->request('GET','/companies/'.rawurlencode($companyId).'/employees',null,$token);
        if($result['ok']) $result['employees']=(array)($result['data']['employees']??[]);
        return $result;
    }
    public function get(string $companyId,string $employeeId,string $token): array {
        return $this->api->request('GET','/companies/'.rawurlencode($companyId).'/employees/'.rawurlencode($employeeId),null,$token);
    }
    public function update(string $companyId,string $employeeId,array $input,string $token): array {
        $keys=['first_name','middle_name','last_name','gender','date_of_birth','national_id','passport_number','nationality','kra_pin','nssf_number','shif_number','nhif_number','employment_date','termination_date','job_title','department_id','employment_type','status'];
        $payload=[]; foreach($keys as $key) $payload[$key]=trim((string)($input[$key]??''));
        foreach(['gender','employment_type','status','kra_pin'] as $key) $payload[$key]=strtoupper($payload[$key]);
        return $this->api->request('PATCH','/companies/'.rawurlencode($companyId).'/employees/'.rawurlencode($employeeId),$payload,$token);
    }
    public function salaryHistory(string $companyId,string $employeeId,string $token): array {
        $result=$this->api->request('GET','/companies/'.rawurlencode($companyId).'/employees/'.rawurlencode($employeeId).'/salary-history',null,$token);
        if($result['ok']) $result['salary_history']=(array)($result['data']['salary_history']??[]);
        return $result;
    }
    public function addSalary(string $companyId,string $employeeId,array $input,string $token): array {
        $payload=[
            'basic_salary'=>str_replace(',','',trim((string)($input['basic_salary']??''))),
            'pay_frequency'=>strtoupper(trim((string)($input['pay_frequency']??''))),
            'effective_from'=>trim((string)($input['effective_from']??'')),
        ];
        return $this->api->request('POST','/companies/'.rawurlencode($companyId).'/employees/'.rawurlencode($employeeId).'/salary-history',$payload,$token);
    }
    public function create(string $companyId,array $input,string $token): array {
        $keys=['employee_number','first_name','middle_name','last_name','gender','date_of_birth','national_id','passport_number','nationality','kra_pin','nssf_number','shif_number','nhif_number','employment_date','job_title','department_id','employment_type','basic_salary','pay_frequency','effective_from'];
        $payload=[]; foreach($keys as $key) $payload[$key]=trim((string)($input[$key]??''));
        foreach(['gender','pay_frequency','employment_type','kra_pin'] as $key) $payload[$key]=strtoupper($payload[$key]);
        return $this->api->request('POST','/companies/'.rawurlencode($companyId).'/employees',$payload,$token);
    }
}
