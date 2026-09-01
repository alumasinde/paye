<?php
declare(strict_types=1);
final class EmployeeController
{
    private function companies(): array { $r=(new CompanyService())->list(auth_access_token()); return $r['ok'] ? $r['companies'] : []; }
    public function index(): void {
        $companies=$this->companies(); $companyId=trim((string)($_GET['company']??''));
        if($companyId===''&&$companies) $companyId=(string)($companies[0]['id']??'');
        $result=$companyId!=='' ? (new EmployeeService())->list($companyId,auth_access_token()) : ['ok'=>true,'employees'=>[]];
        view('employees/index',['title'=>'Employees','layout'=>'app','activeNav'=>'employees','user'=>auth_user(),'companies'=>$companies,'companyId'=>$companyId,'employees'=>$result['ok']?$result['employees']:[],'loadError'=>$result['ok']?'':$result['message']]);
    }
    public function showCreate(): void {
        $companies=$this->companies(); $companyId=trim((string)($_GET['company']??''));
        $old=Session::flash('old');
        $departmentResult=$companyId!=='' ? (new DepartmentService())->list($companyId,auth_access_token()) : ['ok'=>true,'departments'=>[]];
        view('employees/create',['title'=>'Add employee','layout'=>'app','activeNav'=>'employees','user'=>auth_user(),'companies'=>$companies,'companyId'=>$companyId,'departments'=>$departmentResult['ok']?$departmentResult['departments']:[],'old'=>is_array($old)?$old:[]]);
    }
    public function create(): void {
        verify_csrf(); $companyId=trim((string)($_POST['company_id']??'')); $input=$_POST; Session::flash('old',$input);
        foreach(['employee_number','first_name','last_name','employment_date','basic_salary'] as $key) if(trim((string)($input[$key]??''))===''){Session::flash('error','Select a company and complete the required employee and salary details.');redirect('/employees/create?company='.rawurlencode($companyId));}
        $salary=str_replace(',','',(string)$input['basic_salary']);
        if(!is_numeric($salary)||(float)$salary<0){Session::flash('error','Enter a valid basic salary.');redirect('/employees/create?company='.rawurlencode($companyId));}
        $input['basic_salary']=$salary;
        $result=(new EmployeeService())->create($companyId,$input,auth_access_token());
        if(!$result['ok']){Session::flash('error',$result['message']);redirect('/employees/create?company='.rawurlencode($companyId));}
        Session::flash('success','Employee added successfully and is ready for payroll.'); redirect('/employees?company='.rawurlencode($companyId));
    }
}
