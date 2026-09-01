<?php
declare(strict_types=1);
final class DepartmentController
{
    private function companies(): array { $r=(new CompanyService())->list(auth_access_token()); return $r['ok'] ? $r['companies'] : []; }
    public function index(): void {
        $companies=$this->companies(); $companyId=trim((string)($_GET['company']??''));
        if($companyId===''&&$companies)$companyId=(string)($companies[0]['id']??'');
        $r=$companyId!==''?(new DepartmentService())->list($companyId,auth_access_token()):['ok'=>true,'departments'=>[]];
        view('departments/index',['title'=>'Departments','layout'=>'app','activeNav'=>'departments','user'=>auth_user(),'companies'=>$companies,'companyId'=>$companyId,'departments'=>$r['ok']?$r['departments']:[],'loadError'=>$r['ok']?'':$r['message']]);
    }
    public function showCreate(): void {
        $companies=$this->companies();$companyId=trim((string)($_GET['company']??''));$old=Session::flash('old');
        view('departments/create',['title'=>'Add department','layout'=>'app','activeNav'=>'employees','user'=>auth_user(),'companies'=>$companies,'companyId'=>$companyId,'old'=>is_array($old)?$old:[]]);
    }
    public function create(): void {
        verify_csrf();$companyId=trim((string)($_POST['company_id']??''));Session::flash('old',$_POST);
        if($companyId===''||trim((string)($_POST['name']??''))===''){Session::flash('error','Select a company and enter a department name.');redirect('/departments/create?company='.rawurlencode($companyId));}
        $r=(new DepartmentService())->create($companyId,$_POST,auth_access_token());
        if(!$r['ok']){Session::flash('error',$r['message']);redirect('/departments/create?company='.rawurlencode($companyId));}
        Session::flash('success','Department created successfully. You can now assign employees to it.');redirect('/departments?company='.rawurlencode($companyId));
    }
}