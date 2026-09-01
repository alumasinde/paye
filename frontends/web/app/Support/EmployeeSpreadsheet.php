<?php
declare(strict_types=1);

final class EmployeeSpreadsheet
{
    public const COLUMNS = [
        'employee_number','first_name','middle_name','last_name','gender',
        'employment_date','job_title','department_code','employment_type',
        'kra_pin','nssf_number','shif_number','basic_salary','pay_frequency','effective_from',
    ];

    public static function downloadCsv(string $filename, array $headers, iterable $rows): never
    {
        header('Content-Type: text/csv; charset=UTF-8');
        header('Content-Disposition: attachment; filename="' . str_replace('"', '', $filename) . '"');
        header('Cache-Control: no-store');
        echo "\xEF\xBB\xBF";
        $out = fopen('php://output', 'wb');
        fputcsv($out, $headers);
        foreach ($rows as $row) {
            $values = [];
            foreach ($headers as $header) $values[] = (string)($row[$header] ?? '');
            fputcsv($out, $values);
        }
        fclose($out);
        exit;
    }

    public static function templateRows(): array
    {
        return [[
            'employee_number' => 'EMP001',
            'first_name' => 'Jane',
            'middle_name' => '',
            'last_name' => 'Doe',
            'gender' => 'FEMALE',
            'employment_date' => date('Y-m-d'),
            'job_title' => 'Accountant',
            'department_code' => 'FIN',
            'employment_type' => 'PERMANENT',
            'kra_pin' => '',
            'nssf_number' => '',
            'shif_number' => '',
            'basic_salary' => '85000.00',
            'pay_frequency' => 'MONTHLY',
            'effective_from' => date('Y-m-d'),
        ]];
    }

    public static function readUpload(array $file): array
    {
        if (($file['error'] ?? UPLOAD_ERR_NO_FILE) !== UPLOAD_ERR_OK) {
            throw new RuntimeException('Choose a valid CSV or XLSX file.');
        }
        $tmp = (string)($file['tmp_name'] ?? '');
        $size = (int)($file['size'] ?? 0);
        if ($tmp === '' || !is_file($tmp) || $size <= 0) throw new RuntimeException('The uploaded file could not be read.');
        if ($size > 5 * 1024 * 1024) throw new RuntimeException('The import file must be 5 MB or smaller.');

        $extension = strtolower(pathinfo((string)($file['name'] ?? ''), PATHINFO_EXTENSION));
        if ($extension === 'csv') return self::readCsv($tmp);
        if ($extension === 'xlsx') return self::readXlsx($tmp);
        throw new RuntimeException('Use a CSV or XLSX spreadsheet.');
    }

    private static function readCsv(string $path): array
    {
        $handle = fopen($path, 'rb');
        if ($handle === false) throw new RuntimeException('Could not open the CSV file.');
        $rows = [];
        while (($row = fgetcsv($handle)) !== false) $rows[] = $row;
        fclose($handle);
        return self::rowsFromMatrix($rows);
    }

    private static function readXlsx(string $path): array
    {
        if (!class_exists('ZipArchive')) throw new RuntimeException('XLSX import requires the PHP ZipArchive extension.');
        $zip = new ZipArchive();
        if ($zip->open($path) !== true) throw new RuntimeException('The XLSX file could not be opened.');

        $sharedStrings = [];
        $sharedXml = $zip->getFromName('xl/sharedStrings.xml');
        if (is_string($sharedXml)) {
            $xml = self::xml($sharedXml);
            foreach ($xml->si as $si) $sharedStrings[] = trim((string)$si->asXML() === '' ? '' : strip_tags($si->asXML()));
        }

        $sheetPath = 'xl/worksheets/sheet1.xml';
        $sheetXml = $zip->getFromName($sheetPath);
        if (!is_string($sheetXml)) {
            $zip->close();
            throw new RuntimeException('The XLSX file does not contain a first worksheet.');
        }
        $zip->close();

        $xml = self::xml($sheetXml);
        $matrix = [];
        foreach ($xml->sheetData->row as $row) {
            $values = [];
            foreach ($row->c as $cell) {
                $ref = (string)($cell['r'] ?? '');
                $column = preg_replace('/[^A-Z]/', '', $ref) ?: 'A';
                $index = self::columnIndex($column);
                $type = (string)($cell['t'] ?? '');
                $raw = (string)($cell->v ?? '');
                if ($type === 's') $value = $sharedStrings[(int)$raw] ?? '';
                elseif ($type === 'inlineStr') $value = (string)($cell->is->t ?? '');
                else $value = $raw;
                $values[$index] = trim($value);
            }
            if ($values) {
                ksort($values);
                $width = max(array_keys($values)) + 1;
                $matrix[] = array_replace(array_fill(0, $width, ''), $values);
            }
        }
        return self::rowsFromMatrix($matrix);
    }

    private static function xml(string $xml): SimpleXMLElement
    {
        libxml_use_internal_errors(true);
        $previous = libxml_disable_entity_loader(true);
        try {
            $result = simplexml_load_string($xml, SimpleXMLElement::class, LIBXML_NONET | LIBXML_NOCDATA);
        } finally {
            libxml_disable_entity_loader($previous);
        }
        if (!$result instanceof SimpleXMLElement) throw new RuntimeException('The spreadsheet contains invalid XML.');
        return $result;
    }

    private static function columnIndex(string $letters): int
    {
        $value = 0;
        foreach (str_split($letters) as $letter) $value = ($value * 26) + (ord($letter) - 64);
        return $value - 1;
    }

    private static function rowsFromMatrix(array $matrix): array
    {
        if (!$matrix) throw new RuntimeException('The spreadsheet is empty.');
        $headers = array_map(static fn($value): string => self::headerKey((string)$value), array_shift($matrix));
        $missing = array_diff(['employee_number','first_name','last_name','employment_date','basic_salary'], $headers);
        if ($missing) throw new RuntimeException('Missing required columns: ' . implode(', ', $missing) . '.');

        $rows = [];
        foreach ($matrix as $number => $values) {
            if (!array_filter($values, static fn($value) => trim((string)$value) !== '')) continue;
            $row = [];
            foreach ($headers as $index => $header) {
                if ($header !== '') $row[$header] = trim((string)($values[$index] ?? ''));
            }
            $row['_row'] = $number + 2;
            $rows[] = $row;
        }
        if (!$rows) throw new RuntimeException('No employee rows were found in the spreadsheet.');
        if (count($rows) > 1000) throw new RuntimeException('Import a maximum of 1,000 employees at a time.');
        return $rows;
    }

    private static function headerKey(string $value): string
    {
        $key = strtolower(trim($value));
        $key = str_replace([' ','-'], '_', $key);
        $aliases = [
            'employee_no' => 'employee_number',
            'employee_id' => 'employee_number',
            'department' => 'department_code',
            'salary' => 'basic_salary',
            'start_date' => 'employment_date',
        ];
        return $aliases[$key] ?? $key;
    }
}
