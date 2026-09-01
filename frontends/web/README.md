# Payroll web

The payroll web interface is a server-rendered PHP frontend for the Budget254 PAYE API.

Configure the API base URL before deployment:

```env
PAYROLL_API_BASE_URL=https://api.budget254.co.ke/api/v1
```

For local development you can use:

```env
PAYROLL_API_BASE_URL=http://127.0.0.1:8080/api/v1
```

Run it with PHP's built-in server:

```bash
cd frontends/web
PAYROLL_API_BASE_URL=http://127.0.0.1:8080/api/v1 php -S localhost:8088
```
