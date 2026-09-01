<?php
declare(strict_types=1);

final class EmployeeController
{
    private function companies(): array
    {
        $r = (new CompanyService())->list(auth_access_token());
        return $r['ok'] ? $r['companies'] : [];
    }

    private function departments(string $companyId): array
    {
        $r = (new DepartmentService())->list($companyId, auth_access_token());
        return $r['ok'] ? $r['departments'] : [];
    }

    public function index(): void
    {
        $companies = $this->companies();
        $companyId = trim((string)($_GET['company'] ?? ''));
        if ($companyId === '' && $companies) $companyId = (string)($companies[0]['id'] ?? '');
        $result = $companyId !== '' ? (new EmployeeService())->list($companyId, auth_access_token()) : ['ok'=>true,'employees'=>[]];
        view('employees/index', [
            'title'=>'Employees','layout'=>'app','activeNav'=>'employees','user'=>auth_user(),
            'companies'=>$companies,'companyId'=>$companyId,
            'employees'=>$result['ok']?$result['employees']:[],
            'loadError'=>$result['ok']?'':$result['message'],
        ]);
    }

    public function showCreate(): void
    {
        $companies = $this->companies();
        $companyId = trim((string)($_GET['company'] ?? ''));
        $old = Session::flash('old');
        view('employees/create', [
            'title'=>'Add employee','layout'=>'app','activeNav'=>'employees','user'=>auth_user(),
            'companies'=>$companies,'companyId'=>$companyId,
            'departments'=>$companyId !== '' ? $this->departments($companyId) : [],
            'old'=>is_array($old)?$old:[],
        ]);
    }

    public function create(): void
    {
        verify_csrf();
        $companyId = trim((string)($_POST['company_id'] ?? ''));
        $input = $_POST;
        Session::flash('old', $input);
        foreach (['employee_number','first_name','last_name','employment_date','basic_salary'] as $key) {
            if (trim((string)($input[$key] ?? '')) === '') {
                Session::flash('error','Select a company and complete the required employee and salary details.');
                redirect('/employees/create?company=' . rawurlencode($companyId));
            }
        }
        $salary = str_replace(',','',(string)$input['basic_salary']);
        if (!is_numeric($salary) || (float)$salary < 0) {
            Session::flash('error','Enter a valid basic salary.');
            redirect('/employees/create?company=' . rawurlencode($companyId));
        }
        $input['basic_salary'] = $salary;
        $result = (new EmployeeService())->create($companyId, $input, auth_access_token());
        if (!$result['ok']) {
            Session::flash('error',$result['message']);
            redirect('/employees/create?company=' . rawurlencode($companyId));
        }
        Session::flash('success','Employee added successfully and is ready for payroll.');
        redirect('/employees?company=' . rawurlencode($companyId));
    }

    public function downloadTemplate(): void
    {
        EmployeeSpreadsheet::downloadCsv('employee-import-template.csv', EmployeeSpreadsheet::COLUMNS, EmployeeSpreadsheet::templateRows());
    }

    public function export(): void
    {
        $companyId = trim((string)($_GET['company'] ?? ''));
        if ($companyId === '') redirect('/employees');
        $result = (new EmployeeService())->list($companyId, auth_access_token());
        if (!$result['ok']) {
            Session::flash('error', $result['message']);
            redirect('/employees?company=' . rawurlencode($companyId));
        }

        $rows = [];
        foreach ((array)$result['employees'] as $employee) {
            $employee = (array)$employee;
            $salary = (array)($employee['current_salary'] ?? []);
            $rows[] = [
                'employee_number' => (string)($employee['employee_number'] ?? ''),
                'first_name' => (string)($employee['first_name'] ?? ''),
                'middle_name' => (string)($employee['middle_name'] ?? ''),
                'last_name' => (string)($employee['last_name'] ?? ''),
                'gender' => (string)($employee['gender'] ?? ''),
                'employment_date' => (string)($employee['employment_date'] ?? ''),
                'job_title' => (string)($employee['job_title'] ?? ''),
                'department_code' => (string)($employee['department_code'] ?? ''),
                'employment_type' => (string)($employee['employment_type'] ?? ''),
                'kra_pin' => (string)($employee['kra_pin'] ?? ''),
                'nssf_number' => (string)($employee['nssf_number'] ?? ''),
                'shif_number' => (string)($employee['shif_number'] ?? ''),
                'basic_salary' => (string)($salary['basic_salary'] ?? ''),
                'pay_frequency' => (string)($salary['pay_frequency'] ?? ''),
                'effective_from' => (string)($salary['effective_from'] ?? ''),
            ];
        }
        EmployeeSpreadsheet::downloadCsv('employees-' . date('Ymd-His') . '.csv', EmployeeSpreadsheet::COLUMNS, $rows);
    }

    public function showImport(): void
    {
        $companies = $this->companies();
        $companyId = trim((string)($_GET['company'] ?? ''));
        view('employees/import', [
            'title'=>'Import employees','layout'=>'app','activeNav'=>'employees','user'=>auth_user(),
            'companies'=>$companies,'companyId'=>$companyId,
        ]);
    }

    public function previewImport(): void
    {
        verify_csrf();
        $companyId = trim((string)($_POST['company_id'] ?? ''));
        if ($companyId === '') {
            Session::flash('error','Select the company receiving these employees.');
            redirect('/employees/import');
        }

        try {
            $rows = EmployeeSpreadsheet::readUpload($_FILES['spreadsheet'] ?? []);
        } catch (Throwable $e) {
            Session::flash('error', $e->getMessage());
            redirect('/employees/import?company=' . rawurlencode($companyId));
        }

        $departments = $this->departments($companyId);
        $departmentIds = [];
        foreach ($departments as $department) {
            $department = (array)$department;
            $code = strtoupper(trim((string)($department['code'] ?? '')));
            if ($code !== '') $departmentIds[$code] = (string)($department['id'] ?? '');
        }

        $existing = (new EmployeeService())->list($companyId, auth_access_token());
        $existingNumbers = [];
        if ($existing['ok']) foreach ((array)$existing['employees'] as $employee) {
            $existingNumbers[strtoupper(trim((string)($employee['employee_number'] ?? '')))] = true;
        }

        $seen = [];
        $valid = [];
        $invalid = [];
        foreach ($rows as $row) {
            $errors = [];
            $number = strtoupper(trim((string)($row['employee_number'] ?? '')));
            $row['employee_number'] = trim((string)($row['employee_number'] ?? ''));
            foreach (['employee_number','first_name','last_name','employment_date','basic_salary'] as $field) {
                if (trim((string)($row[$field] ?? '')) === '') $errors[] = str_replace('_', ' ', $field) . ' is required';
            }
            if ($number !== '' && isset($seen[$number])) $errors[] = 'duplicate employee number in file';
            if ($number !== '' && isset($existingNumbers[$number])) $errors[] = 'employee number already exists';
            $seen[$number] = true;

            $salary = str_replace(',', '', trim((string)($row['basic_salary'] ?? '')));
            if ($salary === '' || !is_numeric($salary) || (float)$salary < 0) $errors[] = 'basic salary must be a valid positive number';
            $row['basic_salary'] = $salary;

            foreach (['employment_date','effective_from'] as $dateField) {
                $value = trim((string)($row[$dateField] ?? ''));
                if ($value !== '' && !self::validDate($value)) $errors[] = str_replace('_', ' ', $dateField) . ' must use YYYY-MM-DD';
            }
            if (trim((string)($row['effective_from'] ?? '')) === '') $row['effective_from'] = (string)($row['employment_date'] ?? '');

            $row['gender'] = strtoupper(trim((string)($row['gender'] ?? '')));
            if ($row['gender'] !== '' && !in_array($row['gender'], ['MALE','FEMALE','OTHER','PREFER_NOT_TO_SAY'], true)) $errors[] = 'gender is invalid';

            $row['employment_type'] = strtoupper(trim((string)($row['employment_type'] ?? 'PERMANENT')));
            if (!in_array($row['employment_type'], ['PERMANENT','CONTRACT','CASUAL','INTERN','PART_TIME'], true)) $errors[] = 'employment type is invalid';

            $row['pay_frequency'] = strtoupper(trim((string)($row['pay_frequency'] ?? 'MONTHLY')));
            if (!in_array($row['pay_frequency'], ['MONTHLY','WEEKLY','BIWEEKLY'], true)) $errors[] = 'pay frequency is invalid';

            $departmentCode = strtoupper(trim((string)($row['department_code'] ?? '')));
            $row['department_code'] = $departmentCode;
            $row['department_id'] = $departmentCode === '' ? '' : ($departmentIds[$departmentCode] ?? '');
            if ($departmentCode !== '' && $row['department_id'] === '') $errors[] = 'department code was not found for this company';

            if ($errors) {
                $row['_errors'] = $errors;
                $invalid[] = $row;
            } else {
                $valid[] = $row;
            }
        }

        $_SESSION['employee_import'] = [
            'company_id' => $companyId,
            'valid' => $valid,
            'invalid' => $invalid,
            'created_at' => time(),
        ];

        view('employees/import-preview', [
            'title'=>'Review employee import','layout'=>'app','activeNav'=>'employees','user'=>auth_user(),
            'companyId'=>$companyId,'valid'=>$valid,'invalid'=>$invalid,
        ]);
    }

    public function confirmImport(): void
    {
        verify_csrf();
        $batch = (array)($_SESSION['employee_import'] ?? []);
        $companyId = trim((string)($batch['company_id'] ?? ''));
        $valid = (array)($batch['valid'] ?? []);
        if ($companyId === '' || !$valid) {
            Session::flash('error','There is no valid employee import waiting for confirmation.');
            redirect('/employees/import');
        }
        if ((time() - (int)($batch['created_at'] ?? 0)) > 1800) {
            unset($_SESSION['employee_import']);
            Session::flash('error','The import preview expired. Upload the spreadsheet again.');
            redirect('/employees/import?company=' . rawurlencode($companyId));
        }

        $created = 0;
        $failed = [];
        $service = new EmployeeService();
        foreach ($valid as $row) {
            $result = $service->create($companyId, $row, auth_access_token());
            if ($result['ok']) $created++;
            else $failed[] = ['employee_number'=>(string)($row['employee_number'] ?? ''),'message'=>(string)($result['message'] ?? 'Could not create employee.')];
        }
        unset($_SESSION['employee_import']);

        if ($failed) {
            $_SESSION['employee_import_result'] = ['created'=>$created,'failed'=>$failed];
            Session::flash('error', $created . ' employee(s) were created. ' . count($failed) . ' could not be created.');
        } else {
            Session::flash('success', $created . ' employee(s) imported successfully.');
        }
        redirect('/employees?company=' . rawurlencode($companyId));
    }

    public function show(): void
    {
        $companyId = trim((string)($_GET['company'] ?? ''));
        $employeeId = trim((string)($_GET['employee'] ?? ''));
        if ($companyId === '' || $employeeId === '') redirect('/employees');
        $service = new EmployeeService();
        $employeeResult = $service->get($companyId, $employeeId, auth_access_token());
        if (!$employeeResult['ok']) { Session::flash('error',$employeeResult['message']); redirect('/employees?company=' . rawurlencode($companyId)); }
        $historyResult = $service->salaryHistory($companyId, $employeeId, auth_access_token());
        view('employees/show', [
            'title'=>'Employee profile','layout'=>'app','activeNav'=>'employees','user'=>auth_user(),
            'companyId'=>$companyId,'employee'=>(array)$employeeResult['data'],
            'salaryHistory'=>$historyResult['ok']?(array)$historyResult['salary_history']:[],
            'historyError'=>$historyResult['ok']?'':$historyResult['message'],
        ]);
    }

    public function showEdit(): void
    {
        $companyId = trim((string)($_GET['company'] ?? ''));
        $employeeId = trim((string)($_GET['employee'] ?? ''));
        if ($companyId === '' || $employeeId === '') redirect('/employees');
        $result = (new EmployeeService())->get($companyId, $employeeId, auth_access_token());
        if (!$result['ok']) { Session::flash('error',$result['message']); redirect('/employees?company=' . rawurlencode($companyId)); }
        $old = Session::flash('old');
        view('employees/edit', [
            'title'=>'Edit employee','layout'=>'app','activeNav'=>'employees','user'=>auth_user(),
            'companyId'=>$companyId,'employeeId'=>$employeeId,'employee'=>(array)$result['data'],
            'departments'=>$this->departments($companyId),'old'=>is_array($old)?$old:[],
        ]);
    }

    public function update(): void
    {
        verify_csrf();
        $companyId=trim((string)($_POST['company_id']??''));
        $employeeId=trim((string)($_POST['employee_id']??''));
        if ($companyId===''||$employeeId==='') redirect('/employees');
        $input=$_POST;
        foreach(['first_name','last_name','employment_date','status'] as $key) if(trim((string)($input[$key]??''))===''){Session::flash('error','Complete the required employee details.');Session::flash('old',$input);redirect('/employees/edit?company='.rawurlencode($companyId).'&employee='.rawurlencode($employeeId));}
        $result=(new EmployeeService())->update($companyId,$employeeId,$input,auth_access_token());
        if(!$result['ok']){Session::flash('error',$result['message']);Session::flash('old',$input);redirect('/employees/edit?company='.rawurlencode($companyId).'&employee='.rawurlencode($employeeId));}
        Session::flash('success','Employee profile updated.');
        redirect('/employees/show?company='.rawurlencode($companyId).'&employee='.rawurlencode($employeeId));
    }

    public function addSalary(): void
    {
        verify_csrf();
        $companyId=trim((string)($_POST['company_id']??''));
        $employeeId=trim((string)($_POST['employee_id']??''));
        if($companyId===''||$employeeId==='') redirect('/employees');
        $result=(new EmployeeService())->addSalary($companyId,$employeeId,$_POST,auth_access_token());
        if(!$result['ok']) Session::flash('error',$result['message']); else Session::flash('success','Salary change added to the employee history.');
        redirect('/employees/show?company='.rawurlencode($companyId).'&employee='.rawurlencode($employeeId));
    }

    private static function validDate(string $value): bool
    {
        $date = DateTimeImmutable::createFromFormat('!Y-m-d', $value);
        return $date instanceof DateTimeImmutable && $date->format('Y-m-d') === $value;
    }
}
