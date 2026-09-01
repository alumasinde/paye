<?php
declare(strict_types=1);
final class DepartmentService
{
    public function __construct(private readonly ApiService $api = new ApiService()) {}
    public function list(string $companyId,string $token): array {
        $r=$this->api->request('GET','/companies/'.rawurlencode($companyId).'/departments',null,$token);
        if($r['ok']) $r['departments']=(array)($r['data']['departments']??[]);
        return $r;
    }
    public function create(string $companyId,array $input,string $token): array {
        return $this->api->request('POST','/companies/'.rawurlencode($companyId).'/departments',[
            'name'=>trim((string)($input['name']??'')),
            'code'=>strtoupper(trim((string)($input['code']??''))),
            'description'=>trim((string)($input['description']??'')),
        ],$token);
    }
}