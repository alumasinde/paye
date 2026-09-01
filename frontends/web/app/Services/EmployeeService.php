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
    public function create(string $companyId,array $input,string $token): array {
        $keys=['employee_number','first_name','middle_name','last_name','gender','date_of_birth','national_id','passport_number','nationality','kra_pin','nssf_number','shif_number','nhif_number','employment_date','job_title','department','employment_type','basic_salary','pay_frequency','effective_from'];
        $payload=[]; foreach($keys as $key) $payload[$key]=trim((string)($input[$key]??''));
        foreach(['gender','pay_frequency','employment_type','kra_pin'] as $key) $payload[$key]=strtoupper($payload[$key]);
        return $this->api->request('POST','/companies/'.rawurlencode($companyId).'/employees',$payload,$token);
    }
}
