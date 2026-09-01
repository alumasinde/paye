<?php
declare(strict_types=1);

final class AuthService
{
    public function __construct(private readonly ApiService $api = new ApiService()) {}

    public function register(array $input): array
    {
        return $this->api->request('POST', '/auth/register', [
            'email' => trim((string)$input['email']),
            'password' => (string)$input['password'],
            'first_name' => trim((string)$input['first_name']),
            'last_name' => trim((string)$input['last_name']),
        ]);
    }

    public function login(string $email, string $password): array
    {
        return $this->api->request('POST', '/auth/login', [
            'email' => trim($email),
            'password' => $password,
        ]);
    }
}
