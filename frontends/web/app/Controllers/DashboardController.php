<?php
declare(strict_types=1);
final class DashboardController
{
    public function index(): void
    {
        $result = (new CompanyService())->list(auth_access_token());
        $companies = $result['ok'] ? $result['companies'] : [];
        view('dashboard/index', [
            'title'=>'Dashboard','layout'=>'app','activeNav'=>'dashboard','user'=>auth_user(),
            'companies'=>$companies,'companyError'=>$result['ok'] ? '' : $result['message'],
        ]);
    }
}
