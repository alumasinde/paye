<?php
declare(strict_types=1);

final class ApiService
{
    private string $baseUrl;

    public function __construct()
    {
        $this->baseUrl = rtrim(getenv('PAYROLL_API_BASE_URL') ?: 'http://127.0.0.1:8080/api/v1', '/');
    }

    public function request(string $method, string $path, ?array $payload = null, ?string $token = null): array
    {
        $headers = ['Accept: application/json'];
        if ($payload !== null) $headers[] = 'Content-Type: application/json';
        if ($token !== null && $token !== '') $headers[] = 'Authorization: Bearer ' . $token;

        $ch = curl_init($this->baseUrl . '/' . ltrim($path, '/'));
        curl_setopt_array($ch, [
            CURLOPT_CUSTOMREQUEST => strtoupper($method),
            CURLOPT_HTTPHEADER => $headers,
            CURLOPT_RETURNTRANSFER => true,
            CURLOPT_CONNECTTIMEOUT => 8,
            CURLOPT_TIMEOUT => 20,
        ]);
        if ($payload !== null) curl_setopt($ch, CURLOPT_POSTFIELDS, json_encode($payload, JSON_THROW_ON_ERROR));

        $raw = curl_exec($ch);
        $status = (int) curl_getinfo($ch, CURLINFO_RESPONSE_CODE);
        $curlError = curl_error($ch);
        curl_close($ch);

        if ($raw === false || $curlError !== '') return ['ok' => false, 'status' => 0, 'data' => [], 'message' => 'Could not reach the payroll service.'];

        $data = json_decode((string)$raw, true);
        if (!is_array($data)) $data = [];

        return [
            'ok' => $status >= 200 && $status < 300,
            'status' => $status,
            'data' => $data,
            'message' => (string)($data['message'] ?? $data['error']['message'] ?? 'Request could not be completed.'),
        ];
    }
}
