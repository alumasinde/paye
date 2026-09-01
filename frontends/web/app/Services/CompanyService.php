<?php
declare(strict_types=1);

final class CompanyService
{
    public function __construct(private readonly ApiService $api = new ApiService()) {}

    public function list(string $token): array
    {
        $result = $this->api->request('GET', '/companies', null, $token);
        if (!$result['ok']) return $result;
        $result['companies'] = (array)($result['data']['companies'] ?? []);
        return $result;
    }

    public function get(string $companyId, string $token): array
    {
        $result = $this->api->request('GET', '/companies/' . rawurlencode($companyId), null, $token);
        if ($result['ok']) $result['company'] = (array)($result['data'] ?? []);
        return $result;
    }

    public function update(string $companyId, array $input, string $token): array
    {
        $payload = [
            'legal_name' => trim((string)($input['legal_name'] ?? '')),
            'trading_name' => trim((string)($input['trading_name'] ?? '')),
            'email' => strtolower(trim((string)($input['email'] ?? ''))),
            'phone' => trim((string)($input['phone'] ?? '')),
            'country_code' => strtoupper(trim((string)($input['country_code'] ?? 'KE'))),
            'currency_code' => strtoupper(trim((string)($input['currency_code'] ?? 'KES'))),
            'payroll_frequency' => strtoupper(trim((string)($input['payroll_frequency'] ?? 'MONTHLY'))),
            'primary_color' => strtoupper(trim((string)($input['primary_color'] ?? '#15803D'))),
            'secondary_color' => strtoupper(trim((string)($input['secondary_color'] ?? '#166534'))),
        ];
        return $this->api->request('PATCH', '/companies/' . rawurlencode($companyId), $payload, $token);
    }

    public function create(array $input, string $token): array
    {
        $payload = [
            'legal_name' => trim((string)($input['legal_name'] ?? '')),
            'trading_name' => trim((string)($input['trading_name'] ?? '')),
            'kra_pin' => strtoupper(trim((string)($input['kra_pin'] ?? ''))),
            'email' => strtolower(trim((string)($input['email'] ?? ''))),
            'phone' => trim((string)($input['phone'] ?? '')),
            'country_code' => strtoupper(trim((string)($input['country_code'] ?? 'KE'))),
            'currency_code' => strtoupper(trim((string)($input['currency_code'] ?? 'KES'))),
            'payroll_frequency' => strtoupper(trim((string)($input['payroll_frequency'] ?? 'MONTHLY'))),
        ];

        return $this->api->request('POST', '/companies', $payload, $token);
    }
}
